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
	mod.Objects.AddObject(&keys.PrivateKey{})

	return
}
