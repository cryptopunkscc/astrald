package tcp

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

func (mod *Module) Parse(network string, address string) (exonet.Endpoint, error) {
	switch network {
	case "tcp", "inet":
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}

	return tcp.ParseEndpoint(address)
}
