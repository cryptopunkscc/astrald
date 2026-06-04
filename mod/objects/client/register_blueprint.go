package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (client *Client) RegisterBlueprint(ctx *astral.Context, bp *astral.Blueprint) (id *astral.ObjectID, err error) {
	ch, err := client.queryCh(ctx, objects.MethodRegisterBlueprint, nil)
	if err != nil {
		return
	}
	defer ch.Close()

	if err = ch.Send(bp); err != nil {
		return
	}

	err = ch.Switch(channel.Expect(&id), channel.PassErrors)
	return
}

func RegisterBlueprint(ctx *astral.Context, bp *astral.Blueprint) (*astral.ObjectID, error) {
	return Default().RegisterBlueprint(ctx, bp)
}
