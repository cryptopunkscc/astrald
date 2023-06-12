package node

const configName = "node"

type Config struct {
	Identity string   `yaml:"identity"`
	Modules  []string `yaml:"modules"`
}

var defaultConfig = Config{}
