package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

const createUserUrl = "http://localhost:3000"

const colorReset = "\033[0m"
const colorBlue = "\033[34m"
const colorRed = "\033[31m"
const colorGreen = "\033[32m"
const colorCyan = "\033[36m"

const keysFolder = "./keys/"

var orgID int = 1
var userID int = 1

var walletLocation = "wallet"

var channelName string = "sla"
var contractName string = "slasc_bridge"

type userKeys struct {
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

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
						log.Printf(string(colorRed)+"failed to submit transaction: %s\n"+string(colorReset), err)
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

// A wallet can hold multiple identities.
func populateWallet(wallet *gateway.Wallet, orgID int, userID int) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
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

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity(fmt.Sprintf("Org%dMSP", orgID), string(cert), string(key))

	return wallet.Put("appUser", identity)
}

// setDiscoveryAsLocalhost sets the environmental variable DISCOVERY_AS_LOCALHOST
func setDiscoveryAsLocalhost(value bool) error {
	err := os.Setenv("DISCOVERY_AS_LOCALHOST", strconv.FormatBool(value))
	if err != nil {
		return fmt.Errorf("failed to set DISCOVERY_AS_LOCALHOST environment variable: %v", err)
	}
	return nil
}

func userExistsOrCreate(contract *gateway.Contract, name, id string) (bool, string, error) {
	result, err := contract.EvaluateTransaction("UserExists", id)
	if err != nil {
		err = fmt.Errorf(string(colorRed)+"failed to submit transaction: %s\n"+string(colorReset), err)
		return false, "", err
	}
	result_bool, err := strconv.ParseBool(string(result))
	if err != nil {
		err = fmt.Errorf(string(colorRed)+"failed to parse boolean: %s\n"+string(colorReset), err)
		return false, "", err
	}
	if !result_bool {
		postBody, err := json.Marshal(map[string]string{
			"username": id,
		})
		if err != nil {
			err = fmt.Errorf(string(colorRed)+"failed to marshall post request: %s\n"+string(colorReset), err)
			return false, "", err
		}
		responseBody := bytes.NewBuffer(postBody)
		resp, err := http.Post(createUserUrl, "application/json", responseBody)
		if err != nil {
			err = fmt.Errorf(string(colorRed)+"failed to send post request: %s\n"+string(colorReset), err)
			return false, "", err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf(string(colorRed)+"failed to get response body: %s\n"+string(colorReset), err)
			return false, "", err
		}

		// We use interface{} since we don't know the contents of the JSON beforehand
		var responseBodyJSON map[string]interface{}
		err = json.Unmarshal(body, &responseBodyJSON)
		if err != nil {
			err = fmt.Errorf(string(colorRed)+"failed to unmarshall response body: %s\n"+string(colorReset), err)
			return false, "", err
		}
		if responseBodyJSON["success"] == true {
			// get the data of the internal JSON
			data, ok := responseBodyJSON["data"].(map[string]interface{})
			if !ok {
				err = fmt.Errorf(string(colorRed) + "failed to convert interface to struct\n" + string(colorReset))
				return false, "", err
			}
			// convert interface{} to string
			privateKey := fmt.Sprintf("%v", data["privateKey"])
			publicKey := fmt.Sprintf("%v", data["publicKey"])

			publicKeyStripped := splitCertificate(publicKey)
			privateKeyStripped := splitCertificate(privateKey)

			err = saveCertificates(name, privateKeyStripped, publicKeyStripped)
			if err != nil {
				err = fmt.Errorf(string(colorRed)+"failed to save certificates: %s"+string(colorReset), err)
				return false, "", err
			}
			return false, strings.ReplaceAll(publicKeyStripped, "\n", ""), nil
		} else if responseBodyJSON["success"] == false && responseBodyJSON["error"] == "User already exists" {
			return true, "", nil
		} else {
			return false, "", fmt.Errorf(string(colorRed)+"response failure: %v"+string(colorReset), responseBodyJSON["error"])
		}
	}
	return true, "", nil
}

func splitCertificate(certificate string) string {
	certificateSplit := strings.Split(certificate, "-----")
	return strings.Trim(certificateSplit[2], "\n")
}

func saveCertificates(name, privateKey, publicKey string) error {
	data := fmt.Sprintf("PRIVATE KEY:\n---------\n%v\nPUBLIC KEY:\n---------\n%v",
		privateKey, publicKey)
	filename := fmt.Sprintf(keysFolder+"%v.keys", name)
	err := os.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("failed to write keys: %v", err)
	}
	return nil
}
