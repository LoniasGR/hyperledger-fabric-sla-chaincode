package lib

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
