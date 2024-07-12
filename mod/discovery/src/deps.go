package discovery

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/discovery"
)

func (mod *Module) LoadDependencies() error {
	// inject admin command
	if adm, err := core.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(discovery.ModuleName, NewAdmin(mod))
	}

	return nil
}
