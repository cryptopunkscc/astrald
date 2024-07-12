package core

const configName = "node"

type Config struct {
	Identity      string   `yaml:"identity"`
	Modules       []string `yaml:"modules"`
	LogRouteTrace bool     `yaml:"log_route_trace"`
}

var defaultConfig = Config{}
