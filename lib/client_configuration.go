package lib

type Config struct {
	dataFolder         string
	JSONFiles          []string
	contractNamePrefix string
	tlsCertPath        string
	peerEndpoint       string
	gatewayPeer        string
	channelName        string
	chaincodeName      string
	identityEndpoint   string
	consumerGroup      string
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
