package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	// required
	mod.storage, err = modules.Load[storage.Module](mod.node, storage.ModuleName)
	if err != nil {
		return err
	}

	mod.sets, err = modules.Load[sets.Module](mod.node, sets.ModuleName)
	if err != nil {
		return err
	}

	// optional
	mod.fs, _ = modules.Load[fs.Module](mod.node, fs.ModuleName)

	// inject admin command
	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(content.ModuleName, NewAdmin(mod))
	}

	go events.Handle(context.Background(), mod.node.Events(),
		func(event sets.EventMemberUpdate) error {
			if !event.Removed {
				mod.Identify(event.DataID)
			}
			return nil
		})

	mod.identified, err = sets.Open[sets.Basic](mod.sets, content.IdentifiedDataSetName)
	if err != nil {
		mod.identified, err = mod.sets.CreateBasic(content.IdentifiedDataSetName)
		if err != nil {
			return err
		}
	}

	mod.setReady()

	return nil
}
