package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/sig"
)

func (mod *Module) Describe(ctx context.Context, identity id.Identity, opts *desc.Opts) []*desc.Desc {
	endpoints, err := mod.node.Tracker().EndpointsByIdentity(identity)
	if err != nil || len(endpoints) == 0 {
		return nil
	}

	list, _ := sig.MapSlice(endpoints, func(a net.Endpoint) (nodes.Endpoint, error) {
		return nodes.Endpoint{
			Network: a.Network(),
			Address: a.String(),
		}, nil
	})

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data: nodes.Desc{
			Endpoints: list,
		},
	}}
}
