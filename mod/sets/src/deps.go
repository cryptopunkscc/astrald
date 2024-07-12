package sets

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/sets"
)

func (mod *Module) LoadDependencies() error {
	if adm, err := core.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(sets.ModuleName, NewAdmin(mod))
	}

	return nil
}
