package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

func (client *Client) NewTraversal(ctx *astral.Context, target string) (astral.Object, error) {
	ch, err := client.queryCh(ctx, nat.MethodNewTraversal, query.Args{
		"target": target,
	})
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	var obj astral.Object
	err = ch.Switch(
		func(msg *astral.ErrorMessage) error {
			return msg
		},
		func(msg astral.Object) error {
			obj = msg
			return channel.ErrBreak
		},
	)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
