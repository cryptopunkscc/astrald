package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"slices"
)

func (mod *Module) OpOps(ctx *astral.Context, q shell.Query) error {
	term := shell.NewTerminal(q.Accept())
	defer term.Close()

	ops := mod.root.Tree()
	slices.Sort(ops)
	for _, o := range ops {
		term.Printf("%v\n", o)
	}

	return nil
}
