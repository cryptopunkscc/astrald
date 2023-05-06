package gw

type Config struct {
	Gateways []string `yaml:"gateways"`
}

var defaultConfig = Config{}
