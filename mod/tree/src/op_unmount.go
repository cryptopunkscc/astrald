package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opUnmountArgs struct {
	Path string `query:"required"`
	In   string
	Out  string
}

func (mod *Module) OpUnmount(ctx *astral.Context, q *routing.IncomingQuery, args opUnmountArgs) (err error) {
	ch := q.Accept(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	if err := mod.Unmount(args.Path); err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.Ack{})
}
