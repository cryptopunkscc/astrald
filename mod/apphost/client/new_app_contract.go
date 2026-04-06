package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

func NewAppContract(ctx *astral.Context, id *astral.Identity, duration astral.Duration) (*apphost.AppContract, error) {
	return Default().NewAppContract(ctx, id, duration)
}

func (client *Client) NewAppContract(ctx *astral.Context, id *astral.Identity, duration astral.Duration) (contract *apphost.AppContract, err error) {
	ch, err := client.queryCh(ctx, apphost.MethodNewAppContract, query.Args{"ID": id, "Duration": duration})
	if err != nil {
		return
	}
	defer ch.Close()
	err = ch.Switch(channel.Expect(&contract), channel.PassErrors)
	return
}
