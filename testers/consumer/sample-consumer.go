package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {

	// Create the topics that will be used
	topics := make([]string, 2)
	topics[0] = "sla_contracts"
	topics[1] = "sla_violation"

	configFile := lib.ParseArgs()
	conf, err := lib.ReadConfig(*configFile)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	truststore_location_slice := strings.Split(conf["ssl.truststore.location"], "/")
	ca_cert := strings.Join(truststore_location_slice[:len(truststore_location_slice)-1], "/")

	c_sla, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":     conf["bootstrap.servers"],
		"security.protocol":     conf["security.protocol"],
		"ssl.keystore.location": conf["ssl.keystore.location"],
		"ssl.keystore.password": conf["ssl.keystore.password"],
		"ssl.key.password":      conf["ssl.key.password"],
		"ssl.ca.location":       filepath.Join(ca_cert, "server.cer.pem"),
	})
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}

	c_sla.SubscribeTopics([]string{"sla_contracts", "sla_violation"}, nil)

	for {
		msg, err := c_sla.ReadMessage(-1)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}

	c_sla.Close()
}
