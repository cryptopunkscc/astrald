package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opVisibleArgs struct {
	Default *bool `query:"optional"`
}

func (mod *Module) opVisible(ctx astral.Context, env *shell.Env, args opVisibleArgs) (err error) {
	if args.Default == nil {
		return env.Printf("%v%v", mod.visible.Get(), term.Newline{})
	}

	return mod.SetVisible(*args.Default)
}
