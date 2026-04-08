package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

func (client *Client) Bind(ctx *astral.Context) (*channel.Channel, error) {
	ch, err := client.queryCh(ctx, apphost.MethodBind, nil)
	if err != nil {
		return nil, err
	}

	if err = ch.Switch(channel.ExpectAck, channel.PassErrors); err != nil {
		ch.Close()
		return nil, err
	}

	return ch, nil
}

func Bind(ctx *astral.Context) (*channel.Channel, error) {
	return Default().Bind(ctx)
}
