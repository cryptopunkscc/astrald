package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

type opListArgs struct {
	Path string `query:"optional"`
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpList(ctx *astral.Context, q shell.Query, args opListArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	var path = "/"

	if len(args.Path) > 0 {
		path = args.Path
	}

	node, err := tree.Query(ctx, mod.Root(), path, false)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	subs, err := node.Sub(ctx)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	for name := range subs {
		err = ch.Send((*astral.String8)(&name))
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})
}
