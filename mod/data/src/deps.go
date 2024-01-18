package data

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/index"
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

	mod.index, err = modules.Load[index.Module](mod.node, index.ModuleName)
	if err != nil {
		return err
	}

	// optional
	mod.fs, _ = modules.Load[fs.Module](mod.node, fs.ModuleName)

	// inject admin command
	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(data.ModuleName, NewAdmin(mod))
	}

	go events.Handle(context.Background(), mod.node.Events(),
		func(ctx context.Context, event index.EventEntryUpdate) error {
			if event.IndexName == index.LocalNodeUnionName && event.Added {
				mod.Index(event.DataID)
			}
			return nil
		})

	// create an index for identified data
	if _, err = mod.index.IndexInfo(data.IdentifiedDataIndexName); err != nil {
		_, err = mod.index.CreateIndex(data.IdentifiedDataIndexName, index.TypeSet)
		if err != nil {
			return err
		}
	}

	mod.setReady()

	return nil
}
