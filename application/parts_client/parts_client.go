package main

import (
	"crypto/x509"
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
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Config struct {
	dataFolder       string
	dataJSONFile     string
	tlsCertPath      string
	peerEndpoint     string
	gatewayPeer      string
	channelName      string
	chaincodeName    string
	identityEndpoint string
	UserConf         *UserConfig
}

type userCredentials struct {
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"privateKey"`
}

type UserConfig struct {
	Credentials userCredentials `json:"credentials"`
	MspID       string          `json:"mspId"`
	Type        string          `json:"type"`
	Version     int             `json:"version"`
}

func loadConfig() *Config {
	conf := Config{}
	conf.tlsCertPath = "/fabric/tlscacerts/tlsca-signcert.pem"
	conf.peerEndpoint = os.Getenv("fabric_gateway_hostport")
	conf.gatewayPeer = os.Getenv("fabric_gateway_sslHostOverride")
	conf.channelName = os.Getenv("fabric_channel")
	conf.chaincodeName = os.Getenv("fabric_contract")
	conf.identityEndpoint = os.Getenv("identity_endpoint")
	conf.dataFolder = os.Getenv("data_folder")
	conf.dataJSONFile = filepath.Join(conf.dataFolder, "parts.json")

	b, err := os.ReadFile("/fabric/application/wallet/appuser_org3.id")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	var userConf UserConfig
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

	c_parts, err := kafkaUtils.CreateConsumer(*configFile[0], "uc3-consumer-group")
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
		log.Fatalf("failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network := gw.GetNetwork(conf.channelName)
	contract := network.GetContract(conf.chaincodeName)

	log.Println(string(lib.ColorGreen), "--> Submit Transaction: InitLedger, function the connection with the ledger", string(lib.ColorReset))
	_, err = contract.SubmitTransaction("InitLedger")
	if err != nil {
		log.Fatalf("failed to submit transaction: %v", err)
	}

	// Open file for logging incoming json objects
	f, err := lib.OpenJsonFile(conf.dataJSONFile)
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
				log.Fatalf("failed to unmarshal: %s", err)
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
				log.Printf(string(lib.ColorRed)+"failed to submit transaction: %s\n"+string(lib.ColorReset), err)
				continue
			}
			fmt.Println(string(result))
		}
	}
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
	log.Print(conf.UserConf.Credentials.Certificate)
	certificate, err := identity.CertificateFromPEM([]byte(conf.UserConf.Credentials.Certificate))
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(conf.UserConf.MspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign(conf Config) identity.Sign {
	privateKeyPEM := conf.UserConf.Credentials.PrivateKey

	privateKey, err := identity.PrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}
