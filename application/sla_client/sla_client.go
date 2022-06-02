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

var orgID int = 1
var userID int = 1

var walletLocation = "wallet"

var channelName string = "sla"
var contractName string = "slasc_bridge"

func main() {
	// The topics that will be used
	topics := make([]string, 2)
	topics[0] = "sla_contracts"
	topics[1] = "sla_violation"

	log.Println("============ application-golang starts ============")
	err := lib.SetDiscoveryAsLocalhost(true)
	if err != nil {
		log.Fatalf("%v", err)
	}

	configFile := lib.ParseArgs()

	ccpPath := filepath.Join(
		"..",
		"..",
		"organization",
		fmt.Sprintf("connection-org%d.yaml", orgID),
	)

	credPath := filepath.Join(
		"..",
		"..",
		"organization",
		"users",
		fmt.Sprintf("User%d@org%d.example.com", userID, orgID),
		"msp",
	)

	c_sla, err := kafkaUtils.CreateConsumer(*configFile[0], "sla-consumer-group")
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
			msg, err := c_sla.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				log.Fatalf("consumer failed to read: %v", err)
			}

			if *msg.TopicPartition.Topic == topics[0] {
				log.Println(string(msg.Value))
				var sla lib.SLA
				err = json.Unmarshal(msg.Value, &sla)
				if err != nil {
					log.Fatalf("failed to unmarshal: %s", err)
				}
				log.Println(sla)

				exists, providerPubKey, err := lib.UserExistsOrCreate(contract, sla.Details.Provider.Name, sla.Details.Provider.ID)
				if err != nil {
					log.Printf("%v", err)
					continue
				}
				if !exists {
					log.Printf("Provider's public key:\n%v", providerPubKey)
					log.Println(string(lib.ColorGreen), `--> Submit Transaction:
					CreateUser, creates new user with name, ID, publickey and an initial balance`, string(lib.ColorReset))
					_, err := contract.SubmitTransaction("CreateUser",
						sla.Details.Provider.Name, sla.Details.Provider.ID, providerPubKey, "500")
					if err != nil {
						log.Printf(string(lib.ColorCyan)+"failed to submit transaction: %s\n"+string(lib.ColorReset), err)
						continue
					}
				}

				exists, clientPubKey, err := lib.UserExistsOrCreate(contract, sla.Details.Client.Name, sla.Details.Client.ID)
				if err != nil {
					log.Printf("%v", err)
					continue
				}
				if !exists {
					log.Printf("Client's public key:\n%v", clientPubKey)
					log.Println(string(lib.ColorGreen), `--> Submit Transaction:
					CreateUser, creates new user with name, ID, publickey and an initial balance`, string(lib.ColorReset))
					_, err := contract.SubmitTransaction("CreateUser",
						sla.Details.Client.Name, sla.Details.Client.ID, clientPubKey, "500")
					if err != nil {
						log.Printf(string(lib.ColorRed)+"failed to submit transaction: %s\n"+string(lib.ColorReset), err)
						continue
					}
				}

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
				continue
			}
			if *msg.TopicPartition.Topic == topics[1] {
				log.Println(string(msg.Value))
				var v lib.Violation
				err = json.Unmarshal(msg.Value, &v)
				if err != nil {
					log.Fatalf("Unmarshal failed: %s\n", err)
				}
				log.Println(v)

				log.Println(string(lib.ColorGreen), "--> Submit Transaction: SLAViolated, updates contracts details with ID, newStatus", string(lib.ColorReset))
				result, err = contract.SubmitTransaction("SLAViolated", v.SLAID)
				if err != nil {
					log.Printf(string(lib.ColorRed)+"failed to submit transaction: %s\n"+string(lib.ColorReset), err)
					continue
				}
				log.Println(string(result))
				continue
			}
			log.Fatalf("unknown topic %s", *msg.TopicPartition.Topic)
		}
		log.Println("============ application-golang ends ============")
		c_sla.Close()
	}
}
