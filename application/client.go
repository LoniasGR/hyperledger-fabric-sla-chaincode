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
	Customer string `json:"Customer"`
	ID       string `json:"ID"`
	Metric   string `json:"Metric"`
	Provider string `json:"Provider"`
	Status   int    `json:"Status"`
	Value    int    `json:"Value"`
}

type Violation struct {
	ID         string `json:"ID"`
	ContractID string `json:"ContractID"`
	Status     int    `json:"Status"`
}

func main() {
	log.Println("============ application-golang starts ============")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environment variable: %v", err)
	}

	r_SLA := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     "sla",
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	r_Violation := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     "sla_violation",
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	// Cleanup for when the service terminates
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cleanup(r_SLA, r_Violation)
		log.Println("============ application-golang ends ============")
		os.Exit(0)
	}()

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
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
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	contract := network.GetContract("SLA")

	log.Println("--> Submit Transaction: InitLedger, function creates the initial set of assets on the ledger")
	result, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Println(string(result))

	for {
		deadlineError := errors.New("context deadline exceeded")

		waitTime := time.Now().Add(1 * time.Second)
		ctx, cancel := context.WithDeadline(context.Background(), waitTime)
		m, err := r_SLA.ReadMessage(ctx)
		cancel()

		if err == nil {
			log.Println(string(m.Value))
			var s SLA
			err = json.Unmarshal(m.Value, &s)
			if err != nil {
				panic(err.Error())
			}
			log.Println(s)
			log.Println("--> Submit Transaction: CreateContract, creates new contract with ID, customer, metric, provider, value, and status arguments")
			result, err := contract.SubmitTransaction("CreateContract", s.ID, s.Customer, s.Metric, s.Provider, fmt.Sprint(s.Value), fmt.Sprint(s.Status))
			if err != nil {
				log.Fatalf("Failed to submit transaction: %s\n", err)
			}
			fmt.Println(string(result))
		} else if err != nil {
			if err.Error() != deadlineError.Error() {
				log.Fatalf("Reading message failed: %s\n", err)
			}
		}

		waitTime = time.Now().Add(1 * time.Second)
		ctx, cancel = context.WithDeadline(context.Background(), waitTime)
		m, err = r_Violation.ReadMessage(ctx)
		cancel()
		if err == nil {
			log.Println(string(m.Value))

			var v Violation
			err = json.Unmarshal(m.Value, &v)
			if err != nil {
				panic(err.Error())
			}
			log.Println(v)
			log.Println("--> Evaluate Transaction: ContractExists, finds existing contract with ID")
			result, err = contract.EvaluateTransaction("ContractExists", v.ContractID)
			if err != nil {
				log.Fatalf("Failed to evaluate transaction: %s\n", err)
			}
			exists, err := strconv.ParseBool(string(result))
			if err != nil {
				log.Fatalf("Failed to convert result to boolean: %s\n", err)
			}
			if !exists {
				log.Fatalf("Contract with ID: %s does not exist!\n", v.ContractID)
			}

			log.Println("--> Submit Transaction: SLAViolated, updates contracts details with ID, newStatus")
			result, err = contract.SubmitTransaction("SLAViolated", v.ContractID, fmt.Sprint(v.Status))
			if err != nil {
				log.Fatalf("Failed to submit transaction: %s\n", err)
			}
			log.Println(string(result))
		} else if err != nil {
			if err.Error() != deadlineError.Error() {
				log.Fatalf("Reading message failed: %s\n", err)
			}
		}
	}
}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
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

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}

func cleanup(r_SLA *kafka.Reader, r_Violation *kafka.Reader) {
	if err := r_SLA.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
	if err := r_Violation.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}
