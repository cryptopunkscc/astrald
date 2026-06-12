package utp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ nodes.EndpointResolver = &Module{}

// ResolveEndpoints returns the local node's own uTP endpoints; yields an empty
// channel for any other identity.
func (mod *Module) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (_ <-chan *nodes.EndpointWithTTL, err error) {
	if !nodeID.IsEqual(mod.node.Identity()) {
		return sig.ArrayToChan([]*nodes.EndpointWithTTL{}), nil
	}

	return sig.ArrayToChan(mod.endpoints()), nil
}
