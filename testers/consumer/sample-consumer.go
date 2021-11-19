package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
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

func main() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     "SLA",
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cleanup(r)
		os.Exit(0)
	}()

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}

		var s SLA
		err = json.Unmarshal(m.Value, &s)
		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("message at offset %d: %s", m.Offset, string(m.Key))
		fmt.Println(s)

	}
	time.Sleep(time.Second)

}

func cleanup(r *kafka.Reader) {
	if err := r.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}
