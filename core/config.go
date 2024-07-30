package core

const configName = "node"

type Config struct {
	Identity        string   `yaml:"identity"`
	Modules         []string `yaml:"modules"`
	LogRoutingStart bool     `yaml:"log_routing_start"`
}

var defaultConfig = Config{}
