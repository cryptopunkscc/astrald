package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opIndexArgs struct {
	ID   *astral.ObjectID
	Out  string      `query:"optional"`
	Zone astral.Zone `query:"optional"`
}

func (mod *Module) OpIndex(ctx *astral.Context, q *routing.IncomingQuery, args opIndexArgs) (err error) {
	ctx = ctx.WithIdentity(q.Caller())

	ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	err = mod.Index(ctx, args.ID)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}
