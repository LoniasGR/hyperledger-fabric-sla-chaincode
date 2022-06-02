package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/kafkaUtils"
	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

var walletLocation = "wallet"

var orgID int = 1
var userID int = 1

var channelName string = "parts"
var contractName string = "parts"

func main() {

	// The topics that will be used
	topics := make([]string, 1)
	topics[0] = "uc3-dlt"

	log.Println("============ application-golang starts ============")
	err := lib.SetDiscoveryAsLocalhost(true)
	if err != nil {
		log.Fatalf("%v", err)
	}

	configFile := lib.ParseArgs()

	ccpPath := filepath.Join(
		"..",
		"..",
		"organizations",
		"peerOrganizations",
		fmt.Sprintf("org%d.example.com", orgID),
		fmt.Sprintf("connection-org%d.yaml", orgID),
	)

	credPath := filepath.Join(
		"..",
		"..",
		"organizations",
		"peerOrganizations",
		fmt.Sprintf("org%d.example.com", orgID),
		"users",
		fmt.Sprintf("User%d@org%d.example.com", userID, orgID),
		"msp",
	)

	c_parts, err := kafkaUtils.CreateConsumer(*configFile[0], "parts-consumer-group")
	if err != nil {
		log.Fatalf("error in parts_go: %v", err)
	}

	// Subscribe to topic
	err = c_parts.SubscribeTopics(topics, nil)
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
		err = lib.PopulateWallet(wallet, credPath, orgID)
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

	network, err := gw.GetNetwork(channelName)
	if err != nil {
		log.Fatalf("failed to get network: %v", err)
	}
	log.Println("Channel connected")

	contract := network.GetContract(contractName)

	log.Println(string(lib.ColorGreen), "--> Submit Transaction: InitLedger, function the connection with the ledger", string(lib.ColorReset))
	result, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		log.Fatalf("failed to submit transaction: %v", err)
	}
	log.Println(string(result))

	var run bool = true
	for run {
		select {
		case <-sigchan:
			run = false

		default:
			msg, err := c_parts.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				log.Fatalf("consumer failed to read: %v", err)
			}
			log.Println(string(msg.Value))
			var part lib.Part

			err = json.Unmarshal(msg.Value, &part)
			if err != nil {
				log.Fatalf("failed to unmarshal: %s", err)
			}
			log.Println(part)

			log.Println(string(lib.ColorGreen), `--> Submit Transaction:
				CreateContract, creates new contract with ID,
				customer, metric, provider, value, and status arguments`, string(lib.ColorReset))
			result, err = contract.SubmitTransaction("CreateContract",
				string(msg.Value),
			)
			if err != nil {
				log.Printf(string(lib.ColorRed)+"failed to submit transaction: %s\n"+string(lib.ColorReset), err)
				continue
			}
			fmt.Println(string(result))
		}
	}
}
