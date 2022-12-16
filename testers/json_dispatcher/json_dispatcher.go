package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/kafkaUtils"
	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type options_t struct {
	slas       []lib.SLA
	violations []lib.Violation
	vrus       []lib.VRU
	parts      []lib.Part
}

func main() {

	filename := flag.String("json", "", "JSON Input file")
	channel := flag.String("type", "", "Chaincode to be deployed to, one of [sla, violations, vru, parts]")

	configFile := lib.ParseArgs()
	if *channel != "sla" && *channel != "violation" && *channel != "vru" && *channel != "parts" {
		log.Printf("Unknown channel %s", *channel)
		flag.Usage()
		os.Exit(2)
	}

	if _, err := os.Stat(*filename); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("File %s does not exists", *filename)
		os.Exit(2)
	}

	prod, err := kafkaUtils.CreateProducer(*configFile[0])
	if err != nil {
		log.Fatalf("Failure at producer: %v", err)
	}
	defer prod.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range prod.Events() {
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

	jsonFile, err := os.Open(*filename)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := io.ReadAll(jsonFile)

	// Create the topics that will be used
	topic := make([]string, 1)

	var length int
	var options options_t
	switch *channel {
	case "sla":
		topic[0] = "sla_contracts"
		err := json.Unmarshal(byteValue, &options.slas)
		if err != nil {
			var sla lib.SLA
			json.Unmarshal(byteValue, &sla)
			options.slas = append(options.slas, sla)
		}
		length = len(options.slas)
	case "violation":
		topic[0] = "sla_violation"
		json.Unmarshal(byteValue, &options.violations)
		if err != nil {
			var v lib.Violation
			json.Unmarshal(byteValue, &v)
			options.violations = append(options.violations, v)
		}
		length = len(options.violations)
	case "vru":
		topic[0] = "vru_positions"
		json.Unmarshal(byteValue, &options.vrus)
		if err != nil {
			var v lib.VRU
			json.Unmarshal(byteValue, &v)
			options.vrus = append(options.vrus, v)
		}
		length = len(options.vrus)
	case "parts":
		topic[0] = "uc3-dlt"
		json.Unmarshal(byteValue, &options.parts)
		if err != nil {
			var p lib.Part
			json.Unmarshal(byteValue, &p)
			options.parts = append(options.parts, p)
		}
		length = len(options.parts)
	}

	if length <= 0 {
		fmt.Printf("Nothing read from json file. Make sure it's a correctly formatted array or object.")
		os.Exit(1)
	}

	for i := 0; i < length; i++ {
		var err error
		var j []byte
		switch *channel {
		case "sla":
			j, err = json.Marshal(options.slas[i])
		case "violation":
			j, err = json.Marshal(options.violations[i])
		case "vru":
			j, err = json.Marshal(options.vrus[i])
		case "parts":
			j, err = json.Marshal(options.parts[i])
		}

		if err != nil {
			panic(err.Error())
		}

		err = prod.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic[0], Partition: kafka.PartitionAny},
			Value:          j,
		}, nil)
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}
		time.Sleep(1 * time.Second)
	}
	time.Sleep(5 * time.Second)

}
