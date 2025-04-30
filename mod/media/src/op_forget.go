package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
)

type opForgetArgs struct {
	ID  *object.ID
	Out string `query:"optional"`
}

func (mod *Module) OpForget(ctx *astral.Context, q shell.Query, args opForgetArgs) (err error) {
	ctx = ctx.WithIdentity(q.Caller())
	
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	err = mod.Forget(ctx, args.ID)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}
