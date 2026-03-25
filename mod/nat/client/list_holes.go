package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

func (client *Client) ListHoles(ctx *astral.Context, with string) ([]*nat.Hole, error) {
	args := query.Args{}
	if with != "" {
		args["with"] = with
	}

	ch, err := client.queryCh(ctx, nat.MethodListHoles, args)
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	var holes []*nat.Hole

	err = ch.Switch(
		channel.Collect(&holes),
		channel.BreakOnEOS,
		func(msg *astral.ErrorMessage) error {
			return msg
		},
	)

	return holes, err
}
