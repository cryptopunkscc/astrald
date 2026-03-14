package gateway

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ nodes.EndpointResolver = &Module{}

func (mod *Module) ResolveEndpoints(context *astral.Context, nodeID *astral.Identity) (<-chan *nodes.EndpointWithTTL, error) {
	if !nodeID.IsEqual(mod.node.Identity()) {
		// note: we might resolve endpoints if we act as their gateway
		return sig.ArrayToChan([]*nodes.EndpointWithTTL{}), nil
	}

	var endpoints []*nodes.EndpointWithTTL
	for _, gw := range mod.gateways.Clone() {
		endpoints = append(endpoints, nodes.NewEndpointWithTTL(gateway.NewEndpoint(gw, mod.node.Identity()), 7*30*24*time.Hour))
	}

	return sig.ArrayToChan(endpoints), nil
}
