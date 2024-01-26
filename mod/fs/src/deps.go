package fs

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/index"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/modules"
	"time"
)

func (mod *Module) LoadDependencies() error {
	var err error

	// required
	mod.storage, err = modules.Load[storage.Module](mod.node, storage.ModuleName)
	if err != nil {
		return err
	}

	mod.data, err = modules.Load[data.Module](mod.node, data.ModuleName)
	if err != nil {
		return err
	}

	mod.index, err = modules.Load[index.Module](mod.node, index.ModuleName)
	if err != nil {
		return err
	}

	// read only
	mod.storage.AddReader(nameReadOnly, mod.indexer)

	// read write
	mod.storage.AddReader(nameReadWrite, mod.store)
	mod.storage.AddStore(nameReadWrite, mod.store)

	// inject admin command
	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(fs.ModuleName, NewAdmin(mod))
	}

	if mod.data != nil {
		mod.data.AddDescriber(mod)
	}

	// wait for data module to finish preparing
	ctx, cancel := context.WithTimeoutCause(context.Background(), 15*time.Second, errors.New("data module timed out"))
	defer cancel()
	if err := mod.data.Ready(ctx); err != nil {
		return err
	}

	// create our indexes if needed
	if _, err = mod.index.IndexInfo(fs.IndexNameAll); err != nil {
		_, err = mod.index.CreateIndex(fs.IndexNameAll, index.TypeUnion)
		if err != nil {
			return err
		}
		mod.index.AddToUnion(index.LocalNodeUnionName, fs.IndexNameAll)
		mod.index.SetVisible(fs.IndexNameAll, true)
		mod.index.SetDescription(fs.IndexNameAll, "Local filesystem")
	}

	if _, err = mod.index.IndexInfo(nameReadOnly); err != nil {
		_, err = mod.index.CreateIndex(nameReadOnly, index.TypeSet)
		if err != nil {
			return err
		}
		mod.index.AddToUnion(fs.IndexNameAll, nameReadOnly)
	}

	if _, err = mod.index.IndexInfo(nameReadWrite); err != nil {
		_, err = mod.index.CreateIndex(nameReadWrite, index.TypeSet)
		if err != nil {
			return err
		}
		mod.index.AddToUnion(fs.IndexNameAll, nameReadWrite)
	}

	return nil
}
