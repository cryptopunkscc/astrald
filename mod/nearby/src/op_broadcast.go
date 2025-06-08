package nearby

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opBroadcastArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpBroadcast(ctx *astral.Context, q shell.Query, args opBroadcastArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	err = mod.Broadcast()
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}
