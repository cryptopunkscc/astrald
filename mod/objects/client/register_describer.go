package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (client *Client) RegisterDescriber(ctx *astral.Context) error {
	ch, err := client.queryCh(ctx, objects.MethodRegisterDescriber, nil)
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.Switch(channel.ExpectAck, channel.PassErrors, channel.WithContext(ctx))
}

func RegisterDescriber(ctx *astral.Context) error {
	return Default().RegisterDescriber(ctx)
}
