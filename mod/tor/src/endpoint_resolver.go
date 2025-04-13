package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

func (mod *Module) ResolveEndpoints(ctx context.Context, identity *astral.Identity) (endpoints []exonet.Endpoint, err error) {
	if !identity.IsEqual(mod.node.Identity()) {
		return
	}

	if mod.server == nil {
		return
	}
	if mod.server.endpoint == nil || mod.server.endpoint.IsZero() {
		return
	}

	endpoints = append(endpoints, mod.server.endpoint)

	return
}
