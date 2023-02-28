package infra

import (
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
)

const configName = "infra"

type Config struct {
	Networks    []string    `yaml:"networks"`
	Gateways    []string    `yaml:"gateways"`
	StickyNodes []string    `yaml:"sticky_nodes"`
	Inet        inet.Config `yaml:"inet"`
	Tor         tor.Config  `yaml:"tor"`
	Gw          gw.Config   `yaml:"gw"`
}

func (cfg Config) networksContain(network string) bool {
	if len(cfg.Networks) == 0 {
		return true
	}

	for _, n := range cfg.Networks {
		if n == network {
			return true
		}
	}

	return false
}
