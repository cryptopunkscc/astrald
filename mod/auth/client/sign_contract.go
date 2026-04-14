package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

func SignContract(ctx *astral.Context, contract *auth.Contract, as string) (*auth.SignedContract, error) {
	return Default().SignContract(ctx, contract, as)
}

func (c *Client) SignContract(ctx *astral.Context, contract *auth.Contract, as string) (signed *auth.SignedContract, err error) {
	var args any
	if as != "" {
		args = query.Args{"as": as}
	}

	ch, err := c.queryCh(ctx, auth.OpSignContract, args)
	if err != nil {
		return
	}
	defer ch.Close()

	if err = ch.Send(contract); err != nil {
		return
	}

	err = ch.Switch(channel.Expect(&signed), channel.PassErrors)
	return
}
