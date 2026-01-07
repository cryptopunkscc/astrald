package shell

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opOpsArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpOps(ctx *astral.Context, q shell.Query, args opOpsArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	ops := mod.root.Tree()
	slices.Sort(ops)

	for _, o := range ops {
		err = ch.Send((*astral.String8)(&o))
		if err != nil {
			return
		}
	}

	return
}
