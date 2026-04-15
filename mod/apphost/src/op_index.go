package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opIndexArgs struct {
	ID  *astral.ObjectID
	Out string `query:"optional"`
}

func (mod *Module) OpIndex(ctx *astral.Context, q *routing.IncomingQuery, args opIndexArgs) error {
	ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	if err := mod.Index(ctx, args.ID); err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.Ack{})
}
