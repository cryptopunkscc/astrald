package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/user"
)

func (client *Client) SignNodeContract(ctx *astral.Context, contract *user.NodeContract) (signed *user.SignedNodeContract, err error) {
	ch, err := client.queryCh(ctx, user.OpSignNodeContract, nil)
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
