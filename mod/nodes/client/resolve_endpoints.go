package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

func (client *Client) ResolveEndpoints(ctx *astral.Context, identity *astral.Identity) ([]*nodes.ResolvedEndpoint, error) {
	ch, err := client.queryCh(ctx, nodes.MethodResolveEndpoints, query.Args{
		"id": identity,
	})
	if err != nil {
		return nil, err
	}

	var endpoints []*nodes.ResolvedEndpoint
	err = ch.Switch(
		channel.Collect(&endpoints),
		channel.StopOnEOS,
		func(msg *astral.ErrorMessage) error {
			return msg
		},
	)

	return endpoints, err
}
