package network

import (
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
)

type Config struct {
	Inet inet.Config `yaml:"inet"`
	Tor  tor.Config  `yaml:"tor"`
}
