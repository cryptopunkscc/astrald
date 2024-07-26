package sets

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/sets"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(sets.ModuleName, NewAdmin(mod))

	return
}
