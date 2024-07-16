package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

func (mod *Module) Describe(ctx context.Context, identity id.Identity, opts *desc.Opts) []*desc.Desc {
	endpoints := mod.Endpoints(identity)
	if len(endpoints) == 0 {
		return nil
	}

	list, _ := sig.MapSlice(endpoints, func(a exonet.Endpoint) (nodes.Endpoint, error) {
		return nodes.Endpoint{
			Network: a.Network(),
			Address: a.Address(),
		}, nil
	})

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data: nodes.Desc{
			Endpoints: list,
		},
	}}
}
