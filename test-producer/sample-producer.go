package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type SLA struct {
	Customer string `json:"Customer"`
	ID       string `json:"String"`
	Metric   string `json:"Metric"`
	Provider string `json:"Provider"`
	Status   int    `json:"Status"`
	Value    int    `json:"Value"`
}

func main() {
	// to create topics when auto.create.topics.enable='true'
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", "SLA", 0)
	if err != nil {
		panic(err.Error())
	}

	assets := []SLA{
		{ID: "asset1", Customer: "blue", Metric: "Downtime", Provider: "Tomoko", Value: 300, Status: 1},
		{ID: "asset2", Customer: "red", Metric: "Downtime", Provider: "Brad", Value: 400, Status: 1},
		{ID: "asset3", Customer: "green", Metric: "Downtime", Provider: "Jin Soo", Value: 500, Status: 1},
		{ID: "asset4", Customer: "yellow", Metric: "Downtime", Provider: "Max", Value: 600, Status: 1},
		{ID: "asset5", Customer: "black", Metric: "Downtime", Provider: "Adriana", Value: 700, Status: 1},
		{ID: "asset6", Customer: "white", Metric: "Downtime", Provider: "Michel", Value: 800, Status: 1},
	}

	// Set timeout for writing
	for _, asset := range assets {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			panic(err.Error())
		}
		log.Println(assetJSON)
		_, err = conn.WriteMessages(
			kafka.Message{Value: []byte(assetJSON)},
		)
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}
		time.Sleep(10 * time.Second)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}
