package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

func (client *Client) ListPairs(ctx *astral.Context, with string) ([]*nat.TraversedPortPair, error) {
	args := query.Args{}
	if with != "" {
		args["with"] = with
	}

	ch, err := client.queryCh(ctx, nat.MethodListPairs, args)
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	var pairs []*nat.TraversedPortPair

	err = ch.Switch(
		channel.Collect(&pairs),
		channel.StopOnEOS,
		func(msg *astral.ErrorMessage) error {
			return msg
		},
	)

	return pairs, err
}
