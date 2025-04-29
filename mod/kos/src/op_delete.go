package kos

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opDeleteArgs struct {
	Key string
	Out string `query:"optional"`
}

func (mod *Module) OpDelete(ctx *astral.Context, q shell.Query, args opDeleteArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	conn := q.Accept()
	defer conn.Close()

	err = mod.db.Delete(ctx.Identity(), args.Key)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}
