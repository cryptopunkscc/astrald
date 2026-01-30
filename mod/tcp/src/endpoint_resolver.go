package tcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/sig"
)

func (mod *Module) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (_ <-chan exonet.Endpoint, err error) {
	if !nodeID.IsEqual(mod.node.Identity()) {
		return sig.ArrayToChan([]exonet.Endpoint{}), nil
	}

	listen := mod.settings.Listen.Get()
	if listen != nil && !*listen {
		return sig.ArrayToChan([]exonet.Endpoint{}), nil
	}

	return sig.ArrayToChan(mod.endpoints()), nil
}
