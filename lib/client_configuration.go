package lib

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DataFolder         string
	JSONFiles          []string
	ContractNamePrefix string
	TlsCertPath        string
	PeerEndpoint       string
	GatewayPeer        string
	ChannelName        string
	ChaincodeName      string
	IdentityEndpoint   string
	ConsumerGroup      string
	UserConf           *UserConfig
}

type userCredentials struct {
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"privateKey"`
}

type UserConfig struct {
	Credentials userCredentials `json:"credentials"`
	MspID       string          `json:"mspId"`
	Type        string          `json:"type"`
	Version     int             `json:"version"`
}

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

// setDiscoveryAsLocalhost sets the environmental variable DISCOVERY_AS_LOCALHOST
func SetDiscoveryAsLocalhost(value bool) error {
	err := os.Setenv("DISCOVERY_AS_LOCALHOST", strconv.FormatBool(value))
	if err != nil {
		return fmt.Errorf("failed to set DISCOVERY_AS_LOCALHOST environment variable: %v", err)
	}
	return nil
}
