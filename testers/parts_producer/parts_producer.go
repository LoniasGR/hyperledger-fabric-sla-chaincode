package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/kafkaUtils"
	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Create the topics that will be used
	topics := make([]string, 1)
	topics[0] = "uc3-dlt"

	nAssets := flag.Int("a", 5, "Specify how many random assets to produce")

	configFile := lib.ParseArgs()
	p_parts, err := kafkaUtils.CreateProducer(*configFile[0])
	if err != nil {
		log.Fatalf("Failure at producer: %v", err)
	}
	defer p_parts.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range p_parts.Events() {
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
		err = p_parts.Produce(&kafka.Message{
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

func createAssets(nAssets int) []lib.Part {
	quality := []int{0, 1}

	assets := make([]lib.Part, nAssets)
	for i := 0; i < nAssets; i++ {

		cycleTimeBool := rand.Intn(2)

		id := createUid()
		asset := lib.Part{
			Id:           lib.Part_id{Oid: id},
			MA:           "ma-005089",
			Timestamp:    makeTimestampString(),
			Version:      1,
			DocumentType: "T2Bauteil",
			DocumentBody: lib.Part_document_body{
				Start:                 lib.Part_timestamp{Date: makeTimestampInt()},
				Stop:                  lib.Part_timestamp{Date: makeTimestampInt()},
				CycleTime:             0,
				Duration:              28.334,
				ActiveTime:            23.798,
				Quality:               quality[rand.Intn(len(quality))],
				LoadingStop:           lib.Part_timestamp{Date: makeTimestampInt()},
				LoadingTime:           2.478,
				ClampingStarts:        []lib.Part_timestamp{{Date: makeTimestampInt()}},
				ClampingStops:         []lib.Part_timestamp{{Date: makeTimestampInt()}},
				ClampingTimes:         []float32{4.338},
				AdjustingStarts:       []lib.Part_timestamp{{Date: makeTimestampInt()}},
				AdjustingStops:        []lib.Part_timestamp{{Date: makeTimestampInt()}},
				AdjustingTimes:        []float32{4.338},
				ReleasingStarts:       []lib.Part_timestamp{{Date: makeTimestampInt()}},
				ReleasingStops:        []lib.Part_timestamp{{Date: makeTimestampInt()}},
				ReleasingTimes:        []float32{4.338},
				UnloadingStart:        lib.Part_timestamp{Date: makeTimestampInt()},
				UnloadingTime:         2.964,
				Pallet:                1,
				FeedOverride:          91.23438492483591,
				FeedOverrideRapid:     100,
				SpindleOverride:       100,
				ToolChangeOveride:     91.23438492483591,
				SpindleNumber:         1,
				ReleaseLocked:         false,
				NokTBr:                false,
				NokTMo:                false,
				NokRew:                false,
				NokCla:                false,
				NokWpc:                false,
				NokNcP:                false,
				ProductionCondUnavail: false,
				PalletchangeStarts:    []lib.Part_timestamp{{Date: makeTimestampInt()}},
				PalletchangeStops:     []lib.Part_timestamp{{Date: makeTimestampInt()}},
				PalletchangeTimes:     []float32{11.052},
				CarrierID:             1,
				ComponentCode:         "DMC1_+25+4+8+17+56+45",
				ComponentName:         "E_MotorVar2",
				ComponentIdent:        "TestSachnummer2",
				ComponentVersion:      "Index9",
				CycleTimeLoss:         0,
				CycleTimeGain:         0,
				MongoRef: lib.MongRef_s{
					Id:         lib.Part_id{Oid: "6046495b4d1d101a13d5d104"},
					Collection: "Components",
				},
			},
		}
		if cycleTimeBool == 1 {
			asset.DocumentBody.CycleTime = 12.0
		}
		assets[i] = asset
		time.Sleep(1 * time.Second)
	}
	return assets
}

func createUid() string {
	id := uuid.New()
	id_str := strings.Replace(id.String(), "-", "", -1)
	return id_str[:16]
}

func makeTimestampString() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func makeTimestampInt() int {
	return int(time.Now().Unix())
}
