package ether

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type Deps struct {
	Objects objects.Module
	Keys    keys.Module
}

func (mod *Module) LoadDependencies(*astral.Context) (err error) {
	return core.Inject(mod.node, &mod.Deps)
}
