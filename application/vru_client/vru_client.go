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
	conf.DataFolder = os.Getenv("data_folder")
	conf.ConsumerGroup = os.Getenv("consumer_group")
	conf.JSONFiles = make([]string, 1)
	conf.JSONFiles[0] = filepath.Join(conf.DataFolder, "vru.json")

	b, err := os.ReadFile("/fabric/application/wallet/appuser_org2.id")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	var userConf lib.UserConfig
	err = json.Unmarshal(b, &userConf)
	if err != nil {
		log.Fatalf("failed to unmarshal userConf: %v", err)
	}

	conf.UserConf = &userConf
	return &conf
}

func main() {

	// logf, err := os.OpenFile("logs.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatalf("error opening file: %v", err)
	// }
	// defer logf.Close()
	// log.SetOutput(logf)

	conf := loadConfig()

	// The topics that will be used
	topics := make([]string, 1)
	topics[0] = "vru_positions"

	log.Println("============ application-golang starts ============")

	configFile := lib.ParseArgs()

	c_vru, err := lib.CreateConsumer(*configFile[0], conf.ConsumerGroup, "beginning")
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

	clientConnection, err := lib.NewGrpcConnection(*conf)
	if err != nil {
		log.Fatalf("failed to create GRPC connection: %v", err)
	}
	defer clientConnection.Close()

	id, err := lib.NewIdentity(*conf)
	if err != nil {
		log.Fatalf("failed to create identity: %v", err)
	}

	sign, err := lib.NewSign(*conf)
	if err != nil {
		log.Fatalf("failed to create signature: %v", err)
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
		lib.HandleError(err)
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
			msg, err := c_vru.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				log.Printf("consumer failed to read: %v", err)
				continue
			}
			log.Printf("New message received on partition: %v", msg.TopicPartition)
			log.Println(string(msg.Value))
			var vru_slice []lib.VRU

			err = json.Unmarshal(msg.Value, &vru_slice)
			if err != nil {
				var vru lib.VRU
				err = json.Unmarshal(msg.Value, &vru)
				if err != nil {
					log.Printf("failed to unmarshal: %v", err)
					continue
				}
				vru_slice = append(vru_slice, vru)
			}
			log.Println(vru_slice)

			for _, vru := range vru_slice {
				vru_json, err := json.Marshal(vru)
				if err != nil {
					log.Printf("Could not marshall singe vru from slice: %v", err)
				}

				jsonToFile, _ := json.MarshalIndent(vru, "", " ")
				if err = lib.WriteJsonObjectToFile(f, jsonToFile); err != nil {
					log.Printf("%v", err)
				}

				log.Println(string(lib.ColorGreen), `--> Submit Transaction:
				CreateContract, creates new incident with Timestamp,
				and related tram and OBUs incidents`, string(lib.ColorReset))

				result, err := contract.SubmitTransaction("CreateContract",
					string(vru_json),
				)
				if err != nil {
					lib.HandleError(err)
					continue
				}
				log.Println(string(result))
			}
		}
	}
}
