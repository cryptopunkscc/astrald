package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/kcp"
	"github.com/cryptopunkscc/astrald/sig"
)

func (mod *Module) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (_ <-chan exonet.Endpoint, err error) {
	if !nodeID.IsEqual(mod.node.Identity()) {
		return sig.ArrayToChan([]exonet.Endpoint{}), nil
	}

	return sig.ArrayToChan(mod.endpoints()), nil
}

func (mod *Module) endpoints() (list []exonet.Endpoint) {
	ips, _ := mod.IP.LocalIPs()
	for _, tip := range ips {
		e := kcp.Endpoint{
			IP:   tip,
			Port: astral.Uint16(mod.config.ListenPort),
		}

		list = append(list, &e)
	}

	for port := range mod.ephemeralListeners.Clone() {
		for _, tip := range ips {
			e := kcp.Endpoint{
				IP:   tip,
				Port: port,
			}

			list = append(list, &e)
		}
	}

	return list
}
