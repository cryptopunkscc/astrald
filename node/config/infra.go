package config

import (
	iastral "github.com/cryptopunkscc/astrald/infra/astral"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
)

// Infra holds configs for individual infrastructural networks
type Infra struct {
	Inet   inet.Config    `yaml:"inet"`
	Tor    tor.Config     `yaml:"tor"`
	Astral iastral.Config `yaml:"astral"`
}
