package core

const configName = "node"

type Config struct {
	Identity        string    `yaml:"identity,omitempty"`
	Modules         []string  `yaml:"modules,omitempty"`
	LogRoutingStart bool      `yaml:"log_routing_start,omitempty"`
	Log             LogConfig `yaml:"log,omitempty"`
}

type LogConfig struct {
	Level         int  `yaml:"level,omitempty"`
	DisableColors bool `yaml:"disable_colors,omitempty"`
}

var defaultConfig = Config{}
