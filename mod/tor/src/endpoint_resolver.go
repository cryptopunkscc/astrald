package tor

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/sig"
)

func (mod *Module) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (out <-chan exonet.Endpoint, err error) {
	out = sig.ArrayToChan([]exonet.Endpoint{})

	switch {
	case !nodeID.IsEqual(mod.node.Identity()):
		return
	case mod.server == nil:
		return
	case mod.server.endpoint.IsZero():
		return
	}

	return sig.ArrayToChan([]exonet.Endpoint{mod.server.endpoint}), nil
}
