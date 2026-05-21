package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

func (client *Client) UnholdObject(ctx *astral.Context, objectID *astral.ObjectID) error {
	ch, err := client.queryCh(ctx, apphost.MethodUnholdObject, query.Args{
		"id": objectID,
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.Switch(channel.ExpectAck, channel.PassErrors, channel.WithContext(ctx))
}

func UnholdObject(ctx *astral.Context, objectID *astral.ObjectID) error {
	return Default().UnholdObject(ctx, objectID)
}
