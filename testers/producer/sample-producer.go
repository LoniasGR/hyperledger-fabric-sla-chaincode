package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Create the topics that will be used
	topics := make([]string, 2)
	topics[0] = "sla_contracts"
	topics[1] = "sla_violation"

	nAssets := flag.Int("a", 5, "Specify how many random assets to produce")
	nViolations := flag.Int("v", 3, "Specify how many random violations to produce")

	configFile := lib.ParseArgs()
	conf, err := lib.ReadConfig(*configFile)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	truststore_location_slice := strings.Split(conf["ssl.truststore.location"], "/")
	ca_cert := strings.Join(truststore_location_slice[:len(truststore_location_slice)-1], "/")

	p_sla, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":     conf["bootstrap.servers"],
		"security.protocol":     conf["security.protocol"],
		"ssl.keystore.location": conf["ssl.keystore.location"],
		"ssl.keystore.password": conf["ssl.keystore.password"],
		"ssl.key.password":      conf["ssl.key.password"],
		"ssl.ca.location":       filepath.Join(ca_cert, "server.cer.pem"),
	})
	if err != nil {
		log.Fatalf("failed to create producer: %v", err)
	}
	defer p_sla.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range p_sla.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	assets := createAssets(*nAssets)
	// Set timeout for writing
	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			panic(err.Error())
		}
		err = p_sla.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topics[0], Partition: kafka.PartitionAny},
			Value:          assetJSON,
		}, nil)

		if err != nil {
			log.Fatal("failed to write messages: ", err)
		}
		log.Println(asset)
		time.Sleep(1 * time.Second)
	}
	violations := createViolations(*nViolations, *nAssets)

	for _, violation := range violations {
		violationJSON, err := json.Marshal(violation)
		if err != nil {
			panic(err.Error())
		}
		log.Println(violation)
		err = p_sla.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topics[1], Partition: kafka.PartitionAny},
			Value:          violationJSON,
		}, nil)
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}
		time.Sleep(1 * time.Second)
	}
}

func createAssets(nAssets int) []lib.SLA {
	states := []string{"started", "ongoing"} // , "stopped", "deleted"}
	types := []string{"agreement"}
	providers := []lib.Entity{
		{ID: "providerX", Name: "pledger platform"},
		{ID: "providerY", Name: "test tester"},
		{ID: "providerZ", Name: "tester test"}}

	clients := []lib.Entity{
		{ID: "clientX", Name: "client1"},
		{ID: "clientY", Name: "client2"},
		{ID: "clientZ", Name: "client3"},
	}

	nProvider := rand.Intn(len(providers))
	nClient := rand.Intn(len(clients))

	assets := make([]lib.SLA, nAssets)
	for i := 0; i < nAssets; i++ {
		id := fmt.Sprintf("a%d", i)
		name := fmt.Sprintf("Agreement %d", i)
		importance := []lib.Importance{
			{Name: "Warning", Constraint: "> 30"},
			{Name: "Mild", Constraint: "> 30"},
			{Name: "Serious", Constraint: "> 30"},
			{Name: "Sever", Constraint: "> 70"},
			{Name: "Catastrophic", Constraint: "> 70"},
		}
		asset := lib.SLA{
			ID: id, Name: name, State: states[rand.Intn(len(states))],
			Assessment: lib.Assessment{FirstExecution: time.Now().Add(-1000 * time.Hour).Format(time.RFC3339),
				LastExecution: time.Now().Format(time.RFC3339)},
			Details: lib.Detail{
				ID:       id,
				Type:     types[rand.Intn(len(types))],
				Name:     name,
				Provider: providers[nProvider],
				Client:   clients[nClient],
				Creation: time.Now().Format(time.RFC3339),
				Guarantees: []lib.Guarantee{{Name: "TestGuarantee", Constraint: "[test_value] < 0.7", Importance: []lib.Importance{}},
					{Name: "TestGuarantee2", Constraint: "[test_value] < 0.2", Importance: importance}},
				Service: "8",
			},
		}
		assets[i] = asset
	}
	return assets
}

func createViolations(nViolations, nAssets int) []lib.Violation {
	if nAssets == 0 {
		nAssets = 5
	}
	values := []lib.Value{
		{
			Key:      "sum(container_memory_usage_bytes%7Bnamespace='core'%7D){}",
			Value:    rand.Int63n(15270965248),
			Datetime: time.Now().Format(time.RFC3339),
		},
	}

	violations := make([]lib.Violation, nViolations)
	for i := 0; i < nViolations; i++ {
		violation := lib.Violation{
			ID:             fmt.Sprintf("v%d", i),
			SLAID:          fmt.Sprintf("a%d", rand.Intn(nAssets)),
			GuaranteeID:    strconv.Itoa(rand.Intn(100)),
			Datetime:       time.Now().Format(time.RFC3339),
			Constraint:     "[sum(container_memory_usage_bytes%7Bnamespace='core'%7D)] < 30",
			Values:         values,
			ImportanceName: "Catastrophic",
		}
		violations[i] = violation
	}

	return violations
}
