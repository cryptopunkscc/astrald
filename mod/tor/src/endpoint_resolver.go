package tor

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ nodes.EndpointResolver = &Module{}

func (mod *Module) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (out <-chan *nodes.ResolvedEndpoint, err error) {
	out = sig.ArrayToChan([]*nodes.ResolvedEndpoint{})

	listen := mod.settings.Listen.Get()
	if listen != nil && !*listen {
		return sig.ArrayToChan([]*nodes.ResolvedEndpoint{}), nil
	}

	switch {
	case !nodeID.IsEqual(mod.node.Identity()):
		return
	case mod.torServer == nil:
		return
	case mod.torServer.endpoint.IsZero():
		return
	}

	return sig.ArrayToChan([]*nodes.ResolvedEndpoint{
		nodes.NewResolvedEndpoint(mod.torServer.endpoint, 3*30*24*time.Hour),
	}), nil
}
