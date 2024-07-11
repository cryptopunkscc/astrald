package nodes

import (
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.dir, err = modules.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(nodes.ModuleName, NewAdmin(mod))
	}

	mod.keys, _ = modules.Load[keys.Module](mod.node, keys.ModuleName)

	mod.dir.AddDescriber(mod)

	return nil
}
