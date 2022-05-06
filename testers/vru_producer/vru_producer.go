package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Create the topics that will be used
	topics := make([]string, 1)
	topics[0] = "vru_positions"

	nAssets := flag.Int("a", 5, "Specify how many random timestamps to produce")

	configFile := lib.ParseArgs()
	p_vru, err := lib.CreateProducer(*configFile[0])
	if err != nil {
		log.Fatalf("Failure at producer: %v", err)
	}
	defer p_vru.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range p_vru.Events() {
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

	// Set timeout for writing
	for i := 0; i < *nAssets; i++ {
		asset := createAsset()
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			panic(err.Error())
		}
		err = p_vru.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topics[0], Partition: kafka.PartitionAny},
			Value:          assetJSON,
		}, nil)

		if err != nil {
			log.Fatal("failed to write messages: ", err)
		}
		log.Println(asset)
		time.Sleep(1 * time.Second)
	}

}

func randBool() bool {
	return rand.Float32() < 0.5
}

func createLatitude() float64 {
	var min = -90.0
	var max = 90.0
	return min + rand.Float64()*(max-min)
}

func createLongitude() float64 {
	var min = -180.0
	var max = 180.0
	return min + rand.Float64()*(max-min)
}

func createOBU(risk string) lib.OBU_s {
	obu := lib.OBU_s{
		StationID:  rand.Int31(),
		Latitude:   createLatitude(),
		Longditute: createLongitude(),
		Risk:       risk,
	}
	return obu
}

func createTram() lib.Tram_s {
	tram := lib.Tram_s{
		StationID:  rand.Int31(),
		Latitude:   createLatitude(),
		Longditute: createLongitude(),
	}
	return tram
}

func createOBUSlice(nOBUs int, risk []string) []lib.OBU_s {
	obus := make([]lib.OBU_s, nOBUs)
	for i := 0; i < nOBUs; i++ {
		r := risk[rand.Intn(len(risk))]
		obu := createOBU(r)
		obus[i] = obu
	}
	return obus
}

func createAsset() lib.VRU {
	risk := []string{"HIGHRISK", "LOWRISK", "NORISK"}
	nOBUs := rand.Intn(10)
	tramExists := randBool()
	timestamp := time.Now().Unix()

	asset := lib.VRU{
		Timestamp: timestamp,
		OBUs:      createOBUSlice(nOBUs, risk),
	}
	if tramExists {
		asset.Tram = createTram()
	}
	return asset
}

func createAssets(nAssets int) []lib.VRU {

	assets := make([]lib.VRU, nAssets)
	for i := 0; i < nAssets; i++ {
		asset := createAsset()
		assets[i] = asset
	}
	return assets
}
