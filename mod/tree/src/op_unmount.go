package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opUnmountArgs struct {
	Path string
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpUnmount(ctx *astral.Context, q *routing.IncomingQuery, args opUnmountArgs) (err error) {
	ch := channel.New(q.AcceptRaw(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	err = mod.Unmount(args.Path)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}
