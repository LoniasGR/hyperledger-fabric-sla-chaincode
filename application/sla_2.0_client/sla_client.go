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

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/robfig/cron"
)

func loadConfig() *lib.Config {
	conf := lib.Config{}
	conf.TlsCertPath = "/fabric/tlscacerts/tlsca-signcert.pem"
	conf.ContractNamePrefix = os.Getenv("fabric_contract")
	conf.PeerEndpoint = os.Getenv("fabric_gateway_hostport")
	conf.GatewayPeer = os.Getenv("fabric_gateway_sslHostOverride")
	conf.ChannelName = os.Getenv("fabric_channel")
	conf.IdentityEndpoint = os.Getenv("identity_endpoint")
	conf.DataFolder = os.Getenv("data_folder")
	conf.ConsumerGroup = os.Getenv("consumer_group")
	conf.JSONFiles = make([]string, 2)
	conf.JSONFiles[0] = filepath.Join(conf.DataFolder, "sla.json")
	conf.JSONFiles[1] = filepath.Join(conf.DataFolder, "violations.json")

	b, err := os.ReadFile("/fabric/application/wallet/appuser_org4.id")
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
	createKeysFolder(*conf)

	// The topics that will be used
	topics := make([]string, 2)
	topics[0] = "sla_contracts"
	topics[1] = "sla_violation"

	log.Println("============ application-golang starts ============")

	configFile := lib.ParseArgs()

	c_sla, err := lib.CreateConsumer(*configFile[0], conf.ConsumerGroup, "beginning")
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
		panic(err)
	}
	defer gw.Close()

	network := gw.GetNetwork(conf.ChannelName)

	f_sla, err := lib.OpenJsonFile(conf.JSONFiles[0])
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer lib.CloseJsonFile(f_sla)

	f_vio, err := lib.OpenJsonFile(conf.JSONFiles[1])
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
				log.Printf("consumer failed to read: %v", err)
				continue
			}

			if *msg.TopicPartition.Topic == topics[0] {
				log.Println(string(msg.Value))
				var sla lib.SLA
				err = json.Unmarshal(msg.Value, &sla)
				if err != nil {
					log.Printf("failed to unmarshal: %v", err)
					continue
				}
				log.Println(sla)

				// Generate the name of the contract
				contractName := fmt.Sprintf("%v-%v", conf.ContractNamePrefix, sla.ID)

				// Check if the contract exists and otherwise create it
				ok, err := QueryInstalled(4, contractName, *conf)
				if err != nil {
					log.Println(red("%v", err))
					continue
				}
				if !ok {
					err = DeployCC(contractName, 4, *conf)
					if err != nil {
						log.Println(red("%v", err))
						continue
					}
					contract := network.GetContract(contractName)
					// Init ledger
					err = InitLedger(contract)
					if err != nil {
						handleError(err)
						continue
					}
					// Initialize the daily refunding process
					c := cron.New()
					c.AddFunc("@midnight", func() { runRefunds(*contract) })
					c.Start()
				}

				contract := network.GetContract(contractName)

				jsonToFile, _ := json.MarshalIndent(sla, "", " ")
				if err = lib.WriteJsonObjectToFile(f_sla, jsonToFile); err != nil {
					handleError(err)
				}
				log.Println("Creating users and contract")

				_, _, err = UserExistsOrCreate(contract, sla.Details.Provider.Name, 10000, 4, *conf)
				if err != nil {
					handleError(err)
					continue
				}

				_, _, err = UserExistsOrCreate(contract, sla.Details.Client.Name, 10000, 4, *conf)
				if err != nil {
					handleError(err)
					continue
				}

				err = CreateOrUpdateContract(contract, string(msg.Value))
				if err != nil {
					handleError(err)
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
					log.Printf("Unmarshal failed: %v\n", err)
					continue
				}
				log.Println(v)

				jsonToFile, _ := json.MarshalIndent(v, "", " ")
				if err = lib.WriteJsonObjectToFile(f_vio, jsonToFile); err != nil {
					log.Printf("%v", err)
				}
				contractName := fmt.Sprintf("%v-%v", conf.ContractNamePrefix, v.SLAID)
				contract := network.GetContract(contractName)
				if contract == nil {
					log.Printf("failed to find contract %s", contractName)
					continue
				}

				result, err := SLAViolated(contract, string(msg.Value))
				if err != nil {
					handleError(err)
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

func runRefunds(contract client.Contract) error {
	log.Println(string(lib.ColorGreen), `--> Submit Transaction:
	RefundAllSLAs, refund all SLAs`, string(lib.ColorReset))

	_, err := contract.SubmitTransaction("RefundAllSLAs")
	if err != nil {
		return fmt.Errorf(string(lib.ColorRed)+"failed to submit transaction: %w\n"+string(lib.ColorReset), err)
	}
	return nil
}
