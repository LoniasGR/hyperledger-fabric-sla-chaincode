package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

const createUserUrl = "http://localhost:3000"

const keysFolder = "./keys/"

var orgID int = 1
var userID int = 1

var walletLocation = "wallet"

var channelName string = "sla"
var contractName string = "slasc_bridge"

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

	var kafkaConfig = kafka.ConfigMap{
		"bootstrap.servers": conf["bootstrap.servers"],
		"group.id":          "sla-contracts-violations-consumer-group",
		"auto.offset.reset": "beginning",
	}

	if conf["security.protocol"] != "" {
		truststore_location_slice := strings.Split(conf["ssl.truststore.location"], "/")
		ca_cert := strings.Join(truststore_location_slice[:len(truststore_location_slice)-1], "/")

		kafkaConfig.SetKey("ssl.keystore.location", conf["ssl.keystore.location"])
		kafkaConfig.SetKey("security.protocol", conf["security.protocol"])
		kafkaConfig.SetKey("ssl.keystore.password", conf["ssl.keystore.password"])
		kafkaConfig.SetKey("ssl.key.password", conf["ssl.key.password"])
		kafkaConfig.SetKey("ssl.ca.location", filepath.Join(ca_cert, "server.cer.pem"))
	}

	c_sla, err := kafka.NewConsumer(&kafkaConfig)
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

	network, err := gw.GetNetwork(channelName)
	if err != nil {
		log.Fatalf("failed to get network: %v", err)
	}
	log.Println("Channel connected")

	contract := network.GetContract(contractName)

	log.Println(string(colorGreen), "--> Submit Transaction: InitLedger, function the connection with the ledger", string(colorReset))
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

				exists, providerPubKey, err := userExistsOrCreate(contract, sla.Details.Provider.Name, sla.Details.Provider.ID)
				if err != nil {
					log.Printf("%v", err)
					continue
				}
				if !exists {
					log.Printf("Provider's public key:\n%v", providerPubKey)
					log.Println(string(colorGreen), `--> Submit Transaction:
					CreateUser, creates new user with name, ID, publickey and an initial balance`, string(colorReset))
					_, err := contract.SubmitTransaction("CreateUser",
						sla.Details.Provider.Name, sla.Details.Provider.ID, providerPubKey, "500")
					if err != nil {
						log.Printf(string(colorCyan)+"failed to submit transaction: %s\n"+string(colorReset), err)
						continue
					}
				}

				exists, clientPubKey, err := userExistsOrCreate(contract, sla.Details.Client.Name, sla.Details.Client.ID)
				if err != nil {
					log.Printf("%v", err)
					continue
				}
				if !exists {
					log.Printf("Client's public key:\n%v", clientPubKey)
					log.Println(string(colorGreen), `--> Submit Transaction:
					CreateUser, creates new user with name, ID, publickey and an initial balance`, string(colorReset))
					_, err := contract.SubmitTransaction("CreateUser",
						sla.Details.Client.Name, sla.Details.Client.ID, clientPubKey, "500")
					if err != nil {
						log.Printf(string(colorRed)+"failed to submit transaction: %s\n"+string(colorReset), err)
						continue
					}
				}

				log.Println(string(colorGreen), `--> Submit Transaction:
				CreateContract, creates new contract with ID,
				customer, metric, provider, value, and status arguments`, string(colorReset))

				result, err = contract.SubmitTransaction("CreateContract",
					string(msg.Value),
				)
				if err != nil {
					log.Printf(string(colorRed)+"failed to submit transaction: %s\n"+string(colorReset), err)
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

				log.Println(string(colorGreen), "--> Submit Transaction: SLAViolated, updates contracts details with ID, newStatus", string(colorReset))
				result, err = contract.SubmitTransaction("SLAViolated", v.SLAID)
				if err != nil {
					log.Printf(string(colorRed)+"failed to submit transaction: %s\n"+string(colorReset), err)
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
