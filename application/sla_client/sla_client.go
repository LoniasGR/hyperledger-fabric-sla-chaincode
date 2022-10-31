package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/kafkaUtils"
	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/robfig/cron/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Config struct {
	mspID         string
	tlsCertPath   string
	peerEndpoint  string
	gatewayPeer   string
	channelName   string
	chaincodeName string
	UserConf      *UserConfig
}

type userCredentials struct {
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"privateKey"`
}

type UserConfig struct {
	Credentials userCredentials `json:"credentials"`
	MspID       string          `json:"mspId"`
	Type        string          `json:"type"`
	Version     string          `json:"version"`
}

func loadConfig() *Config {
	conf := Config{}
	conf.mspID = os.Getenv("MSP_ID")
	conf.tlsCertPath = "/fabric/tlscacerts/tlsca-signcert.pem"
	conf.peerEndpoint = os.Getenv("fabric_gateway_hostport")
	conf.gatewayPeer = os.Getenv("fabric_gateway_sslHostOverride")
	conf.channelName = os.Getenv("fabric_channel")
	conf.chaincodeName = os.Getenv("fabric_contract")
	b, err := os.ReadFile("/fabric/application/wallet/appuser_org1.id")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	var userConf UserConfig
	err = json.Unmarshal(b, &userConf)
	if err != nil {
		log.Fatalf("failed to unmarsal userConf: %v", err)
	}

	conf.UserConf = &userConf

	return &conf
}

func main() {
	conf := loadConfig()

	// The topics that will be used
	topics := make([]string, 2)
	topics[0] = "sla_contracts"
	topics[1] = "sla_violation"

	log.Println("============ application-golang starts ============")

	configFile := lib.ParseArgs()

	c_sla, err := kafkaUtils.CreateConsumer(*configFile[0], "sla-client-group")
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

	clientConnection := newGrpcConnection(*conf)
	defer clientConnection.Close()

	id := newIdentity(*conf)
	sign := newSign(*conf)

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

	network := gw.GetNetwork(conf.channelName)
	contract := network.GetContract(conf.chaincodeName)

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

				_, _, err := UserExistsOrCreate(contract, sla.Details.Provider.Name, 10000, 1)
				if err != nil {
					log.Printf("%v", err)
					continue
				}

				_, _, err = UserExistsOrCreate(contract, sla.Details.Client.Name, 10000, 1)
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

func runRefunds(contract client.Contract) error {
	log.Println(string(lib.ColorGreen), `--> Submit Transaction:
	RefundAllSLAs, refund all SLAs`, string(lib.ColorReset))

	_, err := contract.SubmitTransaction("RefundAllSLAs")
	if err != nil {
		return fmt.Errorf(string(lib.ColorRed)+"failed to submit transaction: %s\n"+string(lib.ColorReset), err)
	}
	return nil
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection(conf Config) *grpc.ClientConn {
	certificate, err := loadCertificate(conf.tlsCertPath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, "")

	connection, err := grpc.Dial(conf.peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity(conf Config) *identity.X509Identity {
	certificate, err := loadCertificate(conf.UserConf.Credentials.Certificate)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(conf.mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(cert string) (*x509.Certificate, error) {
	return identity.CertificateFromPEM([]byte(cert))
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign(conf Config) identity.Sign {
	privateKeyPEM := conf.UserConf.Credentials.PrivateKey

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}
