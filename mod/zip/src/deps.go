package zip

import (
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/mod/zip"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.content, err = modules.Load[content.Module](mod.node, content.ModuleName)
	if err != nil {
		return err
	}

	mod.storage, err = modules.Load[storage.Module](mod.node, storage.ModuleName)
	if err != nil {
		return err
	}

	mod.shares, err = modules.Load[shares.Module](mod.node, shares.ModuleName)
	if err != nil {
		return err
	}

	mod.sets, err = modules.Load[sets.Module](mod.node, sets.ModuleName)
	if err != nil {
		return err
	}

	// inject admin command
	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(ModuleName, NewAdmin(mod))
	}

	// load set
	mod.archives, err = sets.Open[sets.Union](mod.sets, zip.ArchivesSet)
	if err != nil {
		mod.archives, err = mod.sets.CreateUnion(zip.ArchivesSet)
		if err != nil {
			return err
		}
		mod.sets.Localnode().Add(zip.ArchivesSet)
	}

	mod.content.AddDescriber(mod)
	mod.shares.AddAuthorizer(mod)
	mod.storage.AddReader("mod.zip", mod)

	return nil
}
