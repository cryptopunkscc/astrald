package utp

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/utp"
)

func (mod *Module) Parse(network string, address string) (exonet.Endpoint, error) {
	switch network {
	case "udp":
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}

	return utp.ParseEndpoint(address)
}
