package node

import (
	"os"
)

const configName = "node"

type Config struct {
	Alias   string   `yaml:"alias,omitempty"`
	Modules []string `yaml:"modules"`
}

var defaultConfig = Config{
	Alias: "localnode",
	Modules: []string{
		"admin", "agent", "apphost", "connect", "gateway", "net.keepalive",
		"optimizer", "presence", "net.reflectlink", "roam", "net.tcpfwd",
	},
}

func init() {
	if host, err := os.Hostname(); err == nil {
		defaultConfig.Alias = host
	}
}
