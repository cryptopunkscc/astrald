package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

func (c *Client) IndexContract(ctx *astral.Context, signed *auth.SignedContract) error {
	ch, err := c.queryCh(ctx, auth.OpIndex, nil)
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.Send(signed)
	if err != nil {
		return err
	}

	return ch.Switch(channel.BreakOnEOS, channel.PassErrors)
}

func IndexContract(ctx *astral.Context, signed *auth.SignedContract) error {
	return Default().IndexContract(ctx, signed)
}
