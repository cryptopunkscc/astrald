package config

import (
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
)

// Infra holds configs for individual infrastructural networks
type Infra struct {
	LogLevel int         `yaml:"log_level"`
	Gateways []string    `yaml:"gateways"`
	Inet     inet.Config `yaml:"inet"`
	Tor      tor.Config  `yaml:"tor"`
	Gw       gw.Config   `yaml:"gw"`
}
