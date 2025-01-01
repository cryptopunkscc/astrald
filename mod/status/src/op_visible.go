package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opVisibleArgs struct {
	Default *bool `query:"optional"`
}

func (mod *Module) opVisible(ctx astral.Context, q shell.Query, args opVisibleArgs) (err error) {
	t, err := shell.AcceptTerminal(q)
	if err != nil {
		return err
	}
	defer t.Close()

	if args.Default == nil {
		return t.Printf("%v\n", mod.visible.Get())
	}

	return mod.SetVisible(*args.Default)
}
