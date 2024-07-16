package nodes

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.exonet, err = core.Load[exonet.Module](mod.node, exonet.ModuleName)
	if err != nil {
		return err
	}

	mod.dir, err = core.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	if adm, err := core.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(nodes.ModuleName, NewAdmin(mod))
	}

	mod.keys, _ = core.Load[keys.Module](mod.node, keys.ModuleName)

	mod.dir.AddDescriber(mod)

	return nil
}
