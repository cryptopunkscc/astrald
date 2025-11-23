package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opArgsArgs struct {
	Op  string
	Out string `query:"optional"`
}

func (mod *Module) OpArgs(ctx *astral.Context, q shell.Query, args opArgsArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	op := mod.root.Find(args.Op)
	if op == nil {
		return ch.Write(astral.NewError("op not found"))
	}

	for _, name := range op.ArgNames() {
		err = ch.Write((*astral.String8)(&name))
		if err != nil {
			return
		}
	}

	return
}
