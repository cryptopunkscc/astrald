package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

func (c *Client) SignContract(ctx *astral.Context, contract *auth.Contract) (*auth.SignedContract, error) {
	ch, err := c.queryCh(ctx, auth.MethodSignContract, nil)
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	if err = ch.Send(contract); err != nil {
		return nil, err
	}

	var signed *auth.SignedContract
	err = ch.Switch(channel.Expect(&signed), channel.PassErrors)
	return signed, err
}

func SignContract(ctx *astral.Context, contract *auth.Contract) (*auth.SignedContract, error) {
	return Default().SignContract(ctx, contract)
}
