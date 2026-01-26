package src

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
)

func (mod *Module) LoadDependencies(ctx *astral.Context) (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return err
	}

	mod.Crypto.AddEngine(&Engine{mod: mod})

	return nil
}
