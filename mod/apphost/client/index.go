package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

func Index(ctx *astral.Context, id *astral.ObjectID) error {
	return Default().Index(ctx, id)
}

func (client *Client) Index(ctx *astral.Context, id *astral.ObjectID) error {
	ch, err := client.queryCh(ctx, apphost.MethodIndex, query.Args{"ID": id})
	if err != nil {
		return err
	}
	defer ch.Close()
	return ch.Switch(channel.ExpectAck, channel.PassErrors)
}
