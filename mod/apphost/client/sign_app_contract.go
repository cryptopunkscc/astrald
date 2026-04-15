package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

func SignAppContract(ctx *astral.Context, contract *auth.Contract) (*auth.SignedContract, error) {
	return Default().SignAppContract(ctx, contract)
}

func (client *Client) SignAppContract(ctx *astral.Context, contract *auth.Contract) (signed *auth.SignedContract, err error) {
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
