package keys

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/keys"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(keys.ModuleName, NewAdmin(mod))

	mod.Objects.AddDescriber(mod)
	mod.Objects.AddPrototypes(keys.KeyDesc{})

	return
}
