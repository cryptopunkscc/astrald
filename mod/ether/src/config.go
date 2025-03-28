package ether

const etherUDPPort = 8822

type Config struct {
	UDPPort int `yaml:"udp_port,omitempty"`
}

var defaultConfig = Config{
	UDPPort: etherUDPPort,
}
