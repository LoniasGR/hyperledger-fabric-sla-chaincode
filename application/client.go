/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/segmentio/kafka-go"
)

type SLA struct {
	Customer   string `json:"Customer"`
	ID         string `json:"ID"`
	Metric     string `json:"Metric"`
	Provider   string `json:"Provider"`
	Status     int    `json:"Status"`
	Value      int    `json:"Value"`
	Violations int    `json:"Violations"`
}

type Violation struct {
	ID         string `json:"ID"`
	ContractID string `json:"ContractID"`
}

func main() {
	log.Println("============ application-golang starts ============")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environment variable: %v", err)
	}

	var orgID int = 1
	var walletLocation = fmt.Sprintf("org%d/wallet", orgID)

	var userID int = 1

	var channelName string = "sla"
	var contractName string = "sla_contract"

	var r_SLA *kafka.Reader
	var r_Violation *kafka.Reader

	// Create the topics that will be used (if they don't exist)
	topics := make([]string, 2)
	topics[0] = "sla_contracts"
	topics[1] = "sla_violations"
	createTopic(topics)

	r_SLA = kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     topics[0],
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	r_Violation = kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     topics[1],
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	wallet, err := gateway.NewFileSystemWallet(walletLocation)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists(fmt.Sprintf("appUser%d", userID)) {
		err = populateWallet(wallet, orgID, userID)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		fmt.Sprintf("org%d.example.com", orgID),
		fmt.Sprintf("connection-org%d.yaml", orgID),
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork(channelName)
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	contract := network.GetContract(contractName)

	// Cleanup for when the service terminates
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("============ application-golang ends ============")
		showTransactions(contract)
		cleanup(r_SLA, r_Violation)
		os.Exit(0)
	}()

	log.Println("--> Submit Transaction: InitLedger, function the connection with the ledger")
	result, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		cleanup(r_SLA, r_Violation)
		log.Fatalf("failed to Submit transaction: %v", err)
	}
	log.Println(string(result))

	log.Println("--> Submit Transaction: Mint, function to create tokens")
	result, err = contract.SubmitTransaction("Mint", "100")
	if err != nil {
		cleanup(r_SLA, r_Violation)
		log.Fatalf("failed to submit transaction: %v", err)
	}
	log.Println(string(result))

	for {
		deadlineError := errors.New("context deadline exceeded")

		waitTime := time.Now().Add(500 * time.Millisecond)
		ctx, cancel := context.WithDeadline(context.Background(), waitTime)
		m, err := r_SLA.ReadMessage(ctx)
		cancel()

		// Kinda ugly, maybe replace with a switch?
		if err != nil {
			if err.Error() != deadlineError.Error() {
				cleanup(r_SLA, r_Violation)
				log.Fatalf("reading message failed: %s\n", err)
			}
		} else {
			log.Println(string(m.Value))
			var s SLA
			err = json.Unmarshal(m.Value, &s)
			if err != nil {
				cleanup(r_SLA, r_Violation)
				log.Fatalf("failed to unmarshal: %s\n", err)
			}
			log.Println(s)
			log.Println("--> Submit Transaction: CreateContract, creates new contract with ID, customer, metric, provider, value, and status arguments")
			result, err := contract.SubmitTransaction("CreateContract", s.ID, s.Customer, s.Metric, s.Provider, fmt.Sprint(s.Value), fmt.Sprint(s.Status))
			if err != nil {
				cleanup(r_SLA, r_Violation)
				log.Fatalf("failed to submit transaction: %s\n", err)
			}
			fmt.Println(string(result))
		}

		waitTime = time.Now().Add(500 * time.Millisecond)
		ctx, cancel = context.WithDeadline(context.Background(), waitTime)
		m, err = r_Violation.ReadMessage(ctx)
		cancel()

		if err != nil {
			if err.Error() != deadlineError.Error() {
				cleanup(r_SLA, r_Violation)
				log.Fatalf("reading message failed: %s\n", err)
			}
		} else {
			log.Println(string(m.Value))

			var v Violation
			err = json.Unmarshal(m.Value, &v)
			if err != nil {
				cleanup(r_SLA, r_Violation)
				log.Fatalf("Unmarshal failed: %s\n", err)
			}
			log.Println(v)
			log.Println("--> Evaluate Transaction: ContractExists, finds existing contract with ID")
			result, err = contract.EvaluateTransaction("ContractExists", v.ContractID)
			if err != nil {
				cleanup(r_SLA, r_Violation)
				log.Fatalf("failed to evaluate transaction: %s\n", err)
			}
			exists, err := strconv.ParseBool(string(result))
			if err != nil {
				cleanup(r_SLA, r_Violation)
				log.Fatalf("failed to convert result to boolean: %s\n", err)
			}
			if !exists {
				cleanup(r_SLA, r_Violation)
				log.Fatalf("contract with ID: %s does not exist!\n", v.ContractID)
			}

			log.Println("--> Submit Transaction: SLAViolated, updates contracts details with ID, newStatus")
			result, err = contract.SubmitTransaction("SLAViolated", v.ContractID)
			if err != nil {
				cleanup(r_SLA, r_Violation)
				log.Fatalf("failed to submit transaction: %s\n", err)
			}
			log.Println(string(result))
		}
	}
}

// A wallet can hold multiple identities.
func populateWallet(wallet *gateway.Wallet, orgID int, userID int) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		fmt.Sprintf("org%d.example.com", orgID),
		"users",
		fmt.Sprintf("User%d@org%d.example.com", userID, orgID),
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity(fmt.Sprintf("Org%dMSP", orgNr), string(cert), string(key))

	return wallet.Put(fmt.Sprintf("appUser%d", userID), identity)
}

func showTransactions(contract *gateway.Contract) {
	log.Println("--> Evaluate Transaction: GetAllContracts, finds existing contract with ID")
	result, err := contract.EvaluateTransaction("GetAllContracts")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	log.Println(string(result))
}

func cleanup(r_SLA *kafka.Reader, r_Violation *kafka.Reader) {
	err := r_SLA.Close()
	if err != nil {
		log.Fatal("failed to close reader:", err)
	}
	err = r_Violation.Close()
	if err != nil {
		log.Fatal("failed to close reader:", err)
	}
}

func createTopic(topics []string) {

	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		panic(err.Error())
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		panic(err.Error())
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{}
	for _, topic := range topics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}
	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		panic(err.Error())
	}

}
