package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"slices"
)

type opOpsArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpOps(ctx *astral.Context, q shell.Query, args opOpsArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	ops := mod.root.Tree()
	slices.Sort(ops)

	for _, o := range ops {
		err = ch.Write((*astral.String8)(&o))
		if err != nil {
			return
		}
	}

	return
}
