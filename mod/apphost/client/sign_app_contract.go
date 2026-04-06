package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

func SignAppContract(ctx *astral.Context, contract *apphost.AppContract) (*apphost.SignedAppContract, error) {
	return Default().SignAppContract(ctx, contract)
}

func (client *Client) SignAppContract(ctx *astral.Context, contract *apphost.AppContract) (signed *apphost.SignedAppContract, err error) {
	ch, err := client.queryCh(ctx, apphost.MethodSignAppContract, nil)
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
