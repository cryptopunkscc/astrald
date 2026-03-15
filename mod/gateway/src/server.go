package gateway

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

func (mod *Module) startServers(ctx *astral.Context) {
	for _, addr := range mod.config.Gateway.Listen {
		parts := strings.SplitN(addr, ":", 2)
		if len(parts) != 2 {
			mod.log.Error("invalid listen address: %v", addr)
			continue
		}
		network, address := parts[0], parts[1]
		endpoint, err := mod.Exonet.Parse(network, address)
		if err != nil {
			mod.log.Error("parse listen address %v: %v", addr, err)
			continue
		}

		switch network {
		case "tcp":
			tcpEndpoint, ok := endpoint.(*tcp.Endpoint)
			if !ok {
				mod.log.Error("invalid listen address: %v", addr)
				continue
			}

			mod.log.Logv(1, "start listening on %v", tcpEndpoint)
			if err := mod.TCP.CreateEphemeralListener(ctx, tcpEndpoint.Port, mod.acceptSocketConn); err != nil {
				mod.log.Error("create ephemeral listener on %v: %v", addr, err)
				continue
			}

			mod.listenEndpoints.Set("tcp", tcpEndpoint)
		default:
			mod.log.Error("unsupported gateway socket network: %v", network)
		}
	}
}
