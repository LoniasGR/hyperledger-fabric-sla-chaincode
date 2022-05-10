package kafkaUtils

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// ParseArgs parses the command line arguments and
// returns the config file on success, or exits on error
func ParseArgs() []*string {
	var configFile = flag.String("f", "", "Path to Kafka configuration file")
	var environment = flag.String("e", "prod", "Environment the client is running in. Can be prod or dev")
	flag.Parse()
	if *configFile == "" {
		flag.Usage()
		os.Exit(2) // the same exit code flag.Parse uses
	}

	return []*string{configFile, environment}
}

// ReadConfig reads the file specified by configFile and
// creates a map of key-value pairs that correspond to each
// line of the file. ReadConfig returns the map on success,
// or nil and an error
func ReadConfig(configFile string) (map[string]string, error) {
	m := make(map[string]string)

	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && len(line) != 0 {
			kv := strings.Split(line, "=")
			parameter := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			m[parameter] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return m, nil
}

func GetKafkaConfiguration(configFile string) (kafka.ConfigMap, error) {
	conf, err := ReadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}
	var kafkaConfig = kafka.ConfigMap{
		"bootstrap.servers": conf["bootstrap.servers"],
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

	return kafkaConfig, err
}

func CreateProducer(configFile string) (*kafka.Producer, error) {
	kafkaConfig, err := GetKafkaConfiguration(configFile)
	if err != nil {
		return nil, err
	}

	producer, err := kafka.NewProducer(&kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}
	return producer, nil
}
