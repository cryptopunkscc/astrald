package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opUnmountArgs struct {
	Path string
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpUnmount(ctx *astral.Context, q shell.Query, args opUnmountArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	err = mod.Unmount(args.Path)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}
