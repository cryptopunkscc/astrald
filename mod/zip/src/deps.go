package zip

import (
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/index"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/mod/zip"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.data, err = modules.Load[data.Module](mod.node, data.ModuleName)
	if err != nil {
		return err
	}

	mod.storage, err = modules.Load[storage.Module](mod.node, storage.ModuleName)
	if err != nil {
		return err
	}

	mod.index, err = modules.Load[index.Module](mod.node, index.ModuleName)
	if err != nil {
		return err
	}

	// inject admin command
	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(ModuleName, NewAdmin(mod))
	}

	mod.data.AddDescriber(mod)

	mod.storage.Data().AddReader("mod.zip", mod)
	mod.storage.Access().AddAccessVerifier(mod)

	mod.index.CreateIndex(zip.ArchivesIndexName, index.TypeSet)

	return nil
}
