package vru

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

var walletLocation = "wallet"

var orgID int = 1
var userID int = 1

var channelName string = "vru"
var contractName string = "vru_positions"

func main() {

	// The topics that will be used
	topics := make([]string, 1)
	topics[0] = "vru_positions"

	log.Println("============ application-golang starts ============")
	err := lib.SetDiscoveryAsLocalhost(true)
	if err != nil {
		log.Fatalf("%v", err)
	}

	configFile := lib.ParseArgs()
	conf, err := lib.ReadConfig(*configFile[0])
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

	var kafkaConfig = kafka.ConfigMap{
		"bootstrap.servers": conf["bootstrap.servers"],
		"group.id":          "sla-vru-consumer-group",
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

	c_vru, err := kafka.NewConsumer(&kafkaConfig)
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}

	// Subscribe to topic
	err = c_vru.SubscribeTopics(topics, nil)
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
		err = lib.PopulateWallet(wallet, orgID, userID)
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

	log.Println(string(lib.ColorGreen), "--> Submit Transaction: InitLedger, function the connection with the ledger", string(colorReset))
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
			msg, err := c_vru.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				log.Fatalf("consumer failed to read: %v", err)
			}
			log.Println(string(msg.Value))
			var vru lib.VRU

			err = json.Unmarshal(msg.Value, &vru)
			if err != nil {
				log.Fatalf("failed to unmarshal: %s", err)
			}
			log.Println(vru)

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
