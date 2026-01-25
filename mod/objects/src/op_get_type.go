package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opGetTypeArgs struct {
	ID  *astral.ObjectID
	Out string `query:"optional"`
}

func (mod *Module) OpGetType(ctx *astral.Context, q *ops.Query, args opGetTypeArgs) (err error) {
	ctx = ctx.WithIdentity(q.Caller())

	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	t, err := mod.GetType(ctx, args.ID)
	if err != nil {
		return ch.Send(astral.NewError("unknown type"))
	}

	return ch.Send((*astral.String8)(&t))
}
