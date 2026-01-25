package exonet

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
)

func (mod *Module) LoadDependencies(*astral.Context) (err error) {
	return core.Inject(mod.node, &mod.Deps)
}
