package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opGetTypeArgs struct {
	ID  *astral.ObjectID
	Out string `query:"optional"`
}

func (mod *Module) OpGetType(ctx *astral.Context, q shell.Query, args opGetTypeArgs) (err error) {
	ctx = ctx.WithIdentity(q.Caller())

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	t, err := mod.GetType(ctx, args.ID)
	if err != nil {
		return ch.Write(astral.NewError("unknown type"))
	}

	return ch.Write((*astral.String8)(&t))
}
