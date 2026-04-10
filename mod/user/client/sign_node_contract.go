package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/user"
)

func (client *Client) SignContract(ctx *astral.Context, contract *auth.Contract) (signed *auth.SignedContract, err error) {
	ch, err := client.queryCh(ctx, user.OpSignContract, nil)
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
