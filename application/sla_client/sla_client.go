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
	"github.com/robfig/cron/v3"
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
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		fmt.Sprintf("org%d.example.com", orgID),
		fmt.Sprintf("connection-org%d.yaml", orgID),
	)

	credPath := filepath.Join(
		"..",
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

	c_sla, err := kafkaUtils.CreateConsumer(*configFile[0], "testerino")
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
	_, err = contract.SubmitTransaction("InitLedger")
	if err != nil {
		log.Fatalf("failed to submit transaction: %v", err)
	}

	// Initialize the daily refunding process
	c := cron.New()
	c.AddFunc("@midnight", func() { runRefunds(*contract) })
	c.Start()

	// Inspect the cron job entries' next and previous run times.
	log.Println(c.Entries())

	f_sla, err := lib.OpenJsonFile("slas.json")
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer lib.CloseJsonFile(f_sla)

	f_vio, err := lib.OpenJsonFile("violations.json")
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer lib.CloseJsonFile(f_vio)

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

				jsonToFile, _ := json.MarshalIndent(sla, "", " ")
				if err = lib.WriteJsonObjectToFile(f_sla, jsonToFile); err != nil {
					log.Printf("%v", err)
				}

				_, _, err := lib.UserExistsOrCreate(contract, sla.Details.Provider.Name, 10000, 1)
				if err != nil {
					log.Printf("%v", err)
					continue
				}

				_, _, err = lib.UserExistsOrCreate(contract, sla.Details.Client.Name, 10000, 1)
				if err != nil {
					log.Printf("%v", err)
					continue
				}

				log.Println(string(lib.ColorGreen), `--> Submit Transaction:
				CreateOrUpdateContract, creates new contract or updates existing one with SLA`, string(lib.ColorReset))

				_, err = contract.SubmitTransaction("CreateOrUpdateContract",
					string(msg.Value),
				)
				if err != nil {
					log.Printf(string(lib.ColorRed)+"failed to submit transaction: %s\n"+string(lib.ColorReset), err)
					continue
				}
				log.Println("submitted")
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

				jsonToFile, _ := json.MarshalIndent(v, "", " ")
				if err = lib.WriteJsonObjectToFile(f_vio, jsonToFile); err != nil {
					fmt.Printf("%v", err)
				}

				log.Println(string(lib.ColorGreen), "--> Submit Transaction: SLAViolated, updates contracts details with ID, newStatus", string(lib.ColorReset))
				result, err := contract.SubmitTransaction("SLAViolated", string(msg.Value))
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

func runRefunds(contract gateway.Contract) error {
	log.Println(string(lib.ColorGreen), `--> Submit Transaction:
	RefundAllSLAs, refund all SLAs`, string(lib.ColorReset))

	_, err := contract.SubmitTransaction("RefundAllSLAs")
	if err != nil {
		return fmt.Errorf(string(lib.ColorRed)+"failed to submit transaction: %s\n"+string(lib.ColorReset), err)
	}
	return nil
}
