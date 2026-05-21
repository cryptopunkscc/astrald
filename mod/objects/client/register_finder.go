package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (client *Client) RegisterFinder(ctx *astral.Context) error {
	ch, err := client.queryCh(ctx, objects.MethodRegisterFinder, nil)
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.Switch(channel.ExpectAck, channel.PassErrors, channel.WithContext(ctx))
}

func RegisterFinder(ctx *astral.Context) error {
	return Default().RegisterFinder(ctx)
}
