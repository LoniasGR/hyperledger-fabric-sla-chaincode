/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

func main() {
	log.Println("============ application-golang starts ============")
	err := setDiscoveryAsLocalhost(true)
	if err != nil {
		log.Fatalf("%v", err)
	}

	configFile := lib.ParseArgs()
	conf, err := lib.ReadConfig(*configFile)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	var orgID int = 1
	var userID int = 1

	var walletLocation = "wallet"

	// var channelName string = "sla"
	var contractName string = "sla_contract"

	ccpPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		fmt.Sprintf("org%d.example.com", orgID),
		fmt.Sprintf("connection-org%d.yaml", orgID),
	)

	// The topics that will be used
	topics := make([]string, 2)
	topics[0] = "sla_contracts"
	topics[1] = "sla_violation"

	truststore_location_slice := strings.Split(conf["ssl.truststore.location"], "/")
	ca_cert := strings.Join(truststore_location_slice[:len(truststore_location_slice)-1], "/")

	c_sla, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":       conf["bootstrap.servers"],
		"security.protocol":       conf["security.protocol"],
		"ssl.keystore.location":   conf["ssl.keystore.location"],
		"ssl.keystore.password":   conf["ssl.keystore.password"],
		"ssl.key.password":        conf["ssl.key.password"],
		"ssl.ca.pem":      filepath.Join(ca_cert, "server.cer.pem"),
		"group.id":                "sla",
		"auto.offset.reset":       "earliest",
	})
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}

	// Subscribe to topic
	err = c_sla.SubscribeTopics(topics, nil)
	if err != nil {
		log.Fatalf("failed to connect to topics: %v", err)
	}

	// Cleanup for when the service terminates
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	wallet, err := gateway.NewFileSystemWallet(walletLocation)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet, orgID, userID)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("sla")
	if err != nil {
		log.Fatalf("failed to get network: %v", err)
	}
	log.Println("Channel connected")

	contract := network.GetContract(contractName)

	log.Println("--> Submit Transaction: InitLedger, function the connection with the ledger")
	result, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		// cleanup(r_SLA, r_Violation)
		log.Fatalf("failed to submit transaction: %v", err)
	}
	log.Println(string(result))

	var run bool = true
	for run == true {
		select {
		case <-sigchan:
			run = false

		default:
			msg, err := c_sla.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				log.Fatalf("consumer failed to read: %v", err)
			}

			if msg.TopicPartition.Topic == &topics[0] {
				log.Println(string(msg.Value))
				var sla lib.SLA
				err = json.Unmarshal(msg.Value, &sla)
				if err != nil {
					log.Fatalf("failed to unmarshal: %s", err)
				}
				log.Println(sla)
				log.Println(`--> Submit Transaction: 
				CreateContract, creates new contract with ID, 
				customer, metric, provider, value, and status arguments`)

				result, err := contract.SubmitTransaction("CreateContract",
					string(msg.Value),
				)
				if err != nil {
					log.Fatalf("failed to submit transaction: %s\n", err)
				}
				fmt.Println(string(result))
				continue
			}
			if msg.TopicPartition.Topic == &topics[1] {
				var v lib.Violation
				err = json.Unmarshal(msg.Value, &v)
				if err != nil {
					log.Fatalf("Unmarshal failed: %s\n", err)
				}
				log.Println(v)

				log.Println("--> Submit Transaction: SLAViolated, updates contracts details with ID, newStatus")
				result, err = contract.SubmitTransaction("SLAViolated", v.ContractID)
				if err != nil {
					log.Fatalf("failed to submit transaction: %s\n", err)
				}
				log.Println(string(result))
				continue
			}
			log.Fatalf("unknown topic %s", *msg.TopicPartition.Topic)
		}
		log.Println("============ application-golang ends ============")
		showTransactions(contract)
		c_sla.Close()
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

	identity := gateway.NewX509Identity(fmt.Sprintf("Org%dMSP", orgID), string(cert), string(key))

	return wallet.Put("appUser", identity)
}

func showTransactions(contract *gateway.Contract) {
	log.Println("--> Evaluate Transaction: GetAllContracts, finds existing contract with ID")
	result, err := contract.EvaluateTransaction("GetAllContracts")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	log.Println(string(result))
}

// setDiscoveryAsLocalhost sets the environmental variable DISCOVERY_AS_LOCALHOST
func setDiscoveryAsLocalhost(value bool) error {
	err := os.Setenv("DISCOVERY_AS_LOCALHOST", strconv.FormatBool(value))
	if err != nil {
		return fmt.Errorf("failed to set DISCOVERY_AS_LOCALHOST environment variable: %v", err)
	}
	return nil
}
