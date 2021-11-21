package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strconv"
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

	// Create the topics that will be used
	topics := make([]string, 2)
	topics[0] = "sla"
	topics[1] = "sla_violation"
	createTopic(topics)

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
		err = w.WriteMessages(context.Background(),
			kafka.Message{
				Topic: topics[0],
				Key:   []byte(asset.ID),
				Value: []byte(assetJSON)},
		)
		if err != nil {
			log.Fatal("failed to write messages: ", err)
		}
		log.Println(asset)
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
				Topic: topics[1],
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

func createTopic(topics []string) {

	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		panic(err.Error())
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		panic(err.Error())
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{}
	for _, topic := range topics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}
	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		panic(err.Error())
	}

}
