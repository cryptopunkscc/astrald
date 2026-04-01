package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

func CreateToken(ctx *astral.Context, identity *astral.Identity) (*apphost.AccessToken, error) {
	return Default().CreateToken(ctx, identity)
}

func (client *Client) CreateToken(ctx *astral.Context, identity *astral.Identity) (token *apphost.AccessToken, err error) {
	ch, err := client.queryCh(ctx, apphost.MethodCreateToken, query.Args{"id": identity.String()})
	if err != nil {
		return
	}
	defer ch.Close()
	err = ch.Switch(channel.Expect(&token), channel.PassErrors)
	return
}
