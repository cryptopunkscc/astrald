package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

func (client *Client) Traverse(ctx *astral.Context, target *astral.Identity) (*nat.TraversedPortPair, error) {
	ch, err := client.queryCh(ctx, nat.MethodStartNatTraversal, query.Args{
		"target": target.String(),
	})
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	var pair nat.TraversedPortPair

	err = ch.Switch(
		func(p *nat.TraversedPortPair) error {
			pair = *p
			return channel.ErrBreak
		},
		func(msg *astral.ErrorMessage) error {
			return msg
		},
		channel.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}

	return &pair, nil
}
