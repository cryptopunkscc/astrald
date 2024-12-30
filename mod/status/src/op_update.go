package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

func (mod *Module) opUpdate(ctx astral.Context, env *shell.Env) (err error) {
	return mod.Broadcast()
}
