package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opRemoveFromIndexArgs struct {
	ID  *astral.ObjectID
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpRemoveFromIndex(ctx *astral.Context, query *routing.IncomingQuery, args opRemoveFromIndexArgs) (err error) {
	ch := query.Accept(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	err = mod.RemoveFromIndex(args.ID)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.Ack{})
}
