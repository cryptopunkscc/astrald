package tor

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/sig"
)

func (mod *Module) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (out <-chan exonet.Endpoint, err error) {
	out = sig.ArrayToChan([]exonet.Endpoint{})

	listen := mod.settings.Listen.Get()
	if listen != nil && !*listen {
		return sig.ArrayToChan([]exonet.Endpoint{}), nil
	}

	switch {
	case !nodeID.IsEqual(mod.node.Identity()):
		return
	case mod.torServer == nil:
		return
	case mod.torServer.endpoint.IsZero():
		return
	}

	return sig.ArrayToChan([]exonet.Endpoint{mod.torServer.endpoint}), nil
}
