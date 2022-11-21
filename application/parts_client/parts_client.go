package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func loadConfig() *lib.Config {
	conf := lib.Config{}
	conf.TlsCertPath = "/fabric/tlscacerts/tlsca-signcert.pem"
	conf.PeerEndpoint = os.Getenv("fabric_gateway_hostport")
	conf.GatewayPeer = os.Getenv("fabric_gateway_sslHostOverride")
	conf.ChannelName = os.Getenv("fabric_channel")
	conf.ChaincodeName = os.Getenv("fabric_contract")
	conf.IdentityEndpoint = os.Getenv("identity_endpoint")
	conf.ConsumerGroup = os.Getenv("consumer_group")
	conf.DataFolder = os.Getenv("data_folder")
	conf.JSONFiles = make([]string, 1)
	conf.JSONFiles[0] = filepath.Join(conf.DataFolder, "parts.json")

	b, err := os.ReadFile("/fabric/application/wallet/appuser_org3.id")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	var userConf lib.UserConfig
	err = json.Unmarshal(b, &userConf)
	if err != nil {
		log.Fatalf("failed to unmarsal userConf: %v", err)
	}

	conf.UserConf = &userConf

	log.Print(conf)
	log.Print(userConf)

	return &conf
}

func main() {
	conf := loadConfig()

	// The topics that will be used
	topics := make([]string, 1)
	topics[0] = "uc3-dlt"

	log.Println("============ application-golang starts ============")
	err := lib.SetDiscoveryAsLocalhost(true)
	if err != nil {
		log.Fatalf("%v", err)
	}

	configFile := lib.ParseArgs()

	c_parts, err := lib.CreateConsumer(*configFile[0], conf.ConsumerGroup, "beginning")
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}

	// Subscribe to topic
	err = c_parts.SubscribeTopics(topics, nil)
	if err != nil {
		log.Fatalf("failed to connect to topics: %v", err)
	}

	// Cleanup for when the service terminates
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	clientConnection, err := lib.NewGrpcConnection(*conf)
	if err != nil {
		log.Fatalf("failed to establish GRPC connection: %v", err)
	}
	defer clientConnection.Close()

	id, err := lib.NewIdentity(*conf)
	if err != nil {
		log.Fatalf("failed to create new identity: %v", err)
	}

	sign, err := lib.NewSign(*conf)
	if err != nil {
		log.Fatalf("failed to create new signature: %v", err)

	}

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		log.Fatalf("failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network := gw.GetNetwork(conf.ChannelName)
	contract := network.GetContract(conf.ChaincodeName)

	log.Println(string(lib.ColorGreen), "--> Submit Transaction: InitLedger, function the connection with the ledger", string(lib.ColorReset))
	_, err = contract.SubmitTransaction("InitLedger")
	if err != nil {
		log.Fatalf("failed to submit transaction: %v", err)
	}

	// Open file for logging incoming json objects
	f, err := lib.OpenJsonFile(conf.JSONFiles[0])
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer lib.CloseJsonFile(f)

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
				log.Printf("consumer failed to read: %v", err)
				continue
			}
			// Print object as json
			log.Println(string(msg.Value), '\n')

			var part lib.Part

			// Unmarshal object and print it to stdout
			err = json.Unmarshal(msg.Value, &part)
			if err != nil {
				log.Printf("failed to unmarshal: %v", err)
				continue
			}
			log.Println(part, '\n')

			// Write json object to file
			jsonToFile, _ := json.MarshalIndent(part, "", " ")
			if err = lib.WriteJsonObjectToFile(f, jsonToFile); err != nil {
				log.Printf("%v", err)
			}

			log.Println(string(lib.ColorGreen), `--> Submit Transaction:
				CreateContract, creates new parts entry with ID, Timestamp
				and all Document details`, string(lib.ColorReset))
			result, err := contract.SubmitTransaction("CreateContract",
				string(msg.Value),
			)
			if err != nil {
				log.Printf(string(lib.ColorRed)+"failed to submit transaction: %v\n"+string(lib.ColorReset), err)
				continue
			}
			log.Println(string(result))
		}
	}
}
