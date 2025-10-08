package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opCloseStreamArgs struct {
	ID  astral.Uint64
	Out string `query:"optional"`
}

// OpCloseStream closes a stream with the given id.
func (mod *Module) OpCloseStream(ctx *astral.Context, q shell.Query, args opCloseStreamArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	err = mod.CloseStream(int(args.ID))
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}
