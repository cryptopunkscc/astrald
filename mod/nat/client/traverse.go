package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

func (client *Client) Punch(ctx *astral.Context, target *astral.Identity) (*nat.Hole, error) {
	ch, err := client.queryCh(ctx, nat.MethodPunch, query.Args{
		"target": target.String(),
	})
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	var hole nat.Hole

	err = ch.Switch(
		func(h *nat.Hole) error {
			hole = *h
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

	return &hole, nil
}
