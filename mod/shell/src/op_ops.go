package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"slices"
)

func (mod *Module) OpOps(ctx astral.Context, q shell.Query) error {
	t, err := shell.AcceptTerminal(q)
	if err != nil {
		return err
	}
	defer t.Close()

	ops := mod.root.Tree()
	slices.Sort(ops)
	for _, o := range ops {
		t.Printf("%v\n", o)
	}

	return nil
}
