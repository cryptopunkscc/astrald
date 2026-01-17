package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

type opDeleteArgs struct {
	Path string
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpDelete(ctx *astral.Context, q shell.Query, args opDeleteArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	node, err := tree.Query(ctx, mod.Root(), args.Path, false)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	err = node.Delete(ctx)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}
