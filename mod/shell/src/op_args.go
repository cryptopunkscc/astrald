package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opArgsArgs struct {
	Op  string
	Out string `query:"optional"`
}

func (mod *Module) OpArgs(ctx *astral.Context, q *ops.Query, args opArgsArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	op := mod.root.Find(args.Op)
	if op == nil {
		return ch.Send(astral.NewError("op not found"))
	}

	for _, name := range op.ArgNames() {
		err = ch.Send((*astral.String8)(&name))
		if err != nil {
			return
		}
	}

	return
}
