package fs

import (
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	// required
	mod.storage, err = modules.Load[storage.Module](mod.node, storage.ModuleName)
	if err != nil {
		return err
	}

	mod.data, _ = modules.Load[data.Module](mod.node, data.ModuleName)

	// read only
	mod.storage.Data().AddReader(nameReadOnly, mod.index)
	mod.storage.Data().AddIndex(nameReadOnly, mod.index)

	// read write
	mod.storage.Data().AddReader(nameReadWrite, mod.store)
	mod.storage.Data().AddStore(nameReadWrite, mod.store)
	mod.storage.Data().AddIndex(nameReadWrite, mod.store)

	// inject admin command
	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(fs.ModuleName, NewAdmin(mod))
	}

	if mod.data != nil {
		mod.data.AddDescriber(mod)
	}

	return nil
}
