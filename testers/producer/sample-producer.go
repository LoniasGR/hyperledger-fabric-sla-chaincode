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
	ID       string `json:"ID"`
	Metric   string `json:"Metric"`
	Provider string `json:"Provider"`
	Status   int    `json:"Status"`
	Value    int    `json:"Value"`
}

type Violation struct {
	ID         string `json:"ID"`
	ContractID string `json:"ContractID"`
	Status     int    `json:"Status"`
}

func main() {
	// to create topics when auto.create.topics.enable='true'
	w := &kafka.Writer{
		Addr: kafka.TCP("localhost:9092"),
		// NOTE: When Topic is not defined here, each Message must define it instead.
		Balancer: &kafka.LeastBytes{},
	}

	assets := []SLA{
		{ID: "contract1", Customer: "blue", Metric: "Downtime", Provider: "Tomoko", Value: 300, Status: 1},
		{ID: "contract2", Customer: "red", Metric: "Downtime", Provider: "Brad", Value: 400, Status: 1},
		{ID: "contract3", Customer: "green", Metric: "Downtime", Provider: "Jin Soo", Value: 500, Status: 1},
		{ID: "contract4", Customer: "yellow", Metric: "Downtime", Provider: "Max", Value: 600, Status: 1},
		{ID: "contract5", Customer: "black", Metric: "Downtime", Provider: "Adriana", Value: 700, Status: 1},
		{ID: "contract6", Customer: "white", Metric: "Downtime", Provider: "Michel", Value: 800, Status: 1},
	}

	// Set timeout for writing
	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			panic(err.Error())
		}
		log.Println(asset)
		err = w.WriteMessages(context.Background(),
			kafka.Message{
				Topic: "SLA",
				Key:   []byte(asset.ID),
				Value: []byte(assetJSON)},
		)
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}
		time.Sleep(10 * time.Second)
	}

	violations := []Violation{
		{ID: "violation1", ContractID: "contract1", Status: 2},
		{ID: "violation2", ContractID: "contract3", Status: 2},
		{ID: "violation3", ContractID: "contract5", Status: 2},
	}

	for _, violation := range violations {
		violationJSON, err := json.Marshal(violation)
		if err != nil {
			panic(err.Error())
		}
		log.Println(violation)
		err = w.WriteMessages(context.Background(),
			kafka.Message{
				Topic: "SLAViolations",
				Value: []byte(violationJSON)},
		)
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}
		time.Sleep(10 * time.Second)
	}

	if err := w.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}
