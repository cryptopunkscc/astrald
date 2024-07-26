package setup

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/setup"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(setup.ModuleName, NewAdmin(mod))

	mod.Presence.AddHookAdOut(mod)

	return nil
}
