package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
)

func (client *Client) SetAlias(ctx *astral.Context, identity *astral.Identity, alias string) error {
	ch, err := client.queryCh(ctx, "dir.set_alias", query.Args{
		"id":    identity,
		"alias": alias,
	})
	if err != nil {
		return err
	}

	return ch.Switch(channel.ExpectAck, channel.PassErrors)
}
