package core

const configName = "node"

type Config struct {
	Identity        string   `yaml:"identity"`
	Modules         []string `yaml:"modules"`
	LogRoutingStart bool     `yaml:"log_routing_start"`
	Log             LogConfig
}

type LogConfig struct {
	Level         int  `yaml:"level"`
	DisableColors bool `yaml:"disable_colors"`
}

var defaultConfig = Config{}
