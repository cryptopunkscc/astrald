package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

type opSetArgs struct {
	Path string
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpSet(ctx *astral.Context, q *ops.Query, args opSetArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	node, err := tree.Query(ctx, mod.Root(), args.Path, true)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Handle(ctx, func(object astral.Object) {
		err = node.Set(ctx, object)
		if err != nil {
			ch.Send(astral.NewError(err.Error()))
		} else {
			ch.Send(&astral.Ack{})
		}
	})
}
