package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

func (client *Client) HoldObject(ctx *astral.Context, objectID *astral.ObjectID) error {
	ch, err := client.queryCh(ctx, apphost.MethodHoldObject, query.Args{
		"id": objectID,
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.Switch(channel.ExpectAck, channel.PassErrors, channel.WithContext(ctx))
}

func HoldObject(ctx *astral.Context, objectID *astral.ObjectID) error {
	return Default().HoldObject(ctx, objectID)
}
