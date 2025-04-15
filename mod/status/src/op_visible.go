package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opVisibleArgs struct {
	Arg *bool `query:"optional"`
}

func (ops *Ops) Visible(ctx *astral.Context, q shell.Query, args opVisibleArgs) (err error) {
	t := shell.NewTerminal(q.Accept())
	defer t.Close()

	if args.Arg == nil {
		return t.Printf("%v\n", ops.mod.visible.Get())
	}

	return ops.mod.SetVisible(*args.Arg)
}
