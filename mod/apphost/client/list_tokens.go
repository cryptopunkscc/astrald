package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

func ListTokens(ctx *astral.Context, identity *astral.Identity) ([]*apphost.AccessToken, error) {
	return Default().ListTokens(ctx, identity)
}

func (client *Client) ListTokens(ctx *astral.Context, identity *astral.Identity) (tokens []*apphost.AccessToken, err error) {
	args := query.Args{}
	if !identity.IsZero() {
		args["id"] = identity
	}

	ch, err := client.queryCh(ctx, apphost.MethodListTokens, args)
	if err != nil {
		return
	}
	defer ch.Close()

	err = ch.Switch(
		channel.Collect[*apphost.AccessToken](&tokens),
		channel.PassErrors,
		channel.BreakOnEOS,
	)
	return
}
