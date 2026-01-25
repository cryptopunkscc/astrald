package tree

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

type opGetArgs struct {
	Path   string
	Follow bool   `query:"optional"`
	In     string `query:"optional"`
	Out    string `query:"optional"`
}

func (mod *Module) OpGet(ctx *astral.Context, q *ops.Query, args opGetArgs) (err error) {
	ctx, cancel := ctx.WithCancel()
	defer cancel()

	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	node, err := tree.Query(ctx, mod.Root(), args.Path, false)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	go func() {
		_, _ = ch.Receive()
		cancel()
	}()

	val, err := node.Get(ctx, args.Follow)
	switch {
	case err == nil:

	case errors.Is(err, &tree.ErrNodeHasNoValue{}):
		return ch.Send(&tree.ErrNodeHasNoValue{})

	default:
		return ch.Send(astral.NewError(err.Error()))
	}

	for object := range val {
		ch.Send(object)
	}

	return ch.Send(&astral.EOS{})
}
