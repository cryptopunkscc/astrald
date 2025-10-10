package udp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/sig"
)

// NOTE: Should we expose our UDP endpoints the same way we expose TCP ones?

func (mod *Module) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (_ <-chan exonet.Endpoint, err error) {
	if !nodeID.IsEqual(mod.node.Identity()) {
		return sig.ArrayToChan([]exonet.Endpoint{}), nil
	}

	var all []exonet.Endpoint

	all = append(all, mod.publicEndpoints...)
	all = append(all, mod.localEndpoints()...)

	return sig.ArrayToChan(all), nil
}
