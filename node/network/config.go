package network

import (
	iastral "github.com/cryptopunkscc/astrald/infra/astral"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
)

type Config struct {
	Alias  string         `yaml:"alias"`
	Inet   inet.Config    `yaml:"inet"`
	Tor    tor.Config     `yaml:"tor"`
	Astral iastral.Config `yaml:"astral"`
}
