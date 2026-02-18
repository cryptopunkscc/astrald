package kcp

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/kcp"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ nodes.EndpointResolver = &Module{}

func (mod *Module) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (_ <-chan *nodes.EndpointWithTTL, err error) {
	if !nodeID.IsEqual(mod.node.Identity()) {
		return sig.ArrayToChan([]*nodes.EndpointWithTTL{}), nil
	}

	return sig.ArrayToChan(mod.endpoints()), nil
}

func (mod *Module) endpoints() (list []*nodes.EndpointWithTTL) {
	ips, _ := mod.IP.LocalIPs()
	for _, tip := range ips {
		e := &kcp.Endpoint{
			IP:   tip,
			Port: astral.Uint16(mod.config.ListenPort),
		}

		list = append(list, nodes.NewEndpointWithTTL(e, 7*24*time.Hour))
	}

	for port := range mod.ephemeralListeners.Clone() {
		for _, tip := range ips {
			e := &kcp.Endpoint{
				IP:   tip,
				Port: port,
			}

			list = append(list, nodes.NewEndpointWithTTL(e, 7*24*time.Hour))
		}
	}

	return list
}
