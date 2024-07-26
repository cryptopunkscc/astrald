package presence

import (
	"github.com/cryptopunkscc/astrald/core"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(ModuleName, NewAdmin(mod))

	return nil
}
