package objects

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
)

func (client *Client) Push(ctx *astral.Context, object astral.Object) error {
	ch, err := client.queryCh(ctx, "objects.push", nil)
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.Send(object)
	if err != nil {
		return err
	}

	return ch.Switch(
		func(result *astral.Bool) error {
			if *result {
				return channel.ErrBreak
			}
			return errors.New("rejected")
		}, channel.PassErrors, channel.WithContext(ctx))
}

func Push(ctx *astral.Context, object astral.Object) error {
	return Default().Push(ctx, object)
}
