package fs

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/resources"
	"os"
	"path/filepath"
	"time"
)

func (mod *Module) LoadDependencies() error {
	var err error

	// required
	mod.storage, err = modules.Load[storage.Module](mod.node, storage.ModuleName)
	if err != nil {
		return err
	}

	mod.content, err = modules.Load[content.Module](mod.node, content.ModuleName)
	if err != nil {
		return err
	}

	mod.sets, err = modules.Load[sets.Module](mod.node, sets.ModuleName)
	if err != nil {
		return err
	}

	// read only
	mod.storage.AddReader(fs.ReadOnlySetName, mod.indexer)

	// read write
	mod.storage.AddReader(fs.ReadWriteSetName, mod.store)
	mod.storage.AddStore(fs.ReadWriteSetName, mod.store)

	// memory
	if mod.mem != nil {
		mod.storage.AddReader(fs.MemorySetName, mod.mem)
		mod.storage.AddStore(fs.MemorySetName, mod.mem)
	}

	// inject admin command
	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(fs.ModuleName, NewAdmin(mod))
	}

	if mod.content != nil {
		mod.content.AddDescriber(mod)
	}

	// wait for data module to finish preparing
	ctx, cancel := context.WithTimeoutCause(context.Background(), 15*time.Second, errors.New("data module timed out"))
	defer cancel()
	if err := mod.content.Ready(ctx); err != nil {
		return err
	}

	// create our indexes if needed
	if _, err = mod.sets.SetInfo(fs.AllSetName); err != nil {
		_, err = mod.sets.CreateSet(fs.AllSetName, sets.TypeUnion)
		if err != nil {
			return err
		}
		mod.sets.AddToUnion(sets.LocalNodeSet, fs.AllSetName)
		mod.sets.SetVisible(fs.AllSetName, true)
		mod.sets.SetDescription(fs.AllSetName, "Local filesystem")
	}

	if _, err = mod.sets.SetInfo(fs.ReadOnlySetName); err != nil {
		_, err = mod.sets.CreateSet(fs.ReadOnlySetName, sets.TypeSet)
		if err != nil {
			return err
		}
		mod.sets.AddToUnion(fs.AllSetName, fs.ReadOnlySetName)
	}

	if _, err = mod.sets.SetInfo(fs.ReadWriteSetName); err != nil {
		_, err = mod.sets.CreateSet(fs.ReadWriteSetName, sets.TypeSet)
		if err != nil {
			return err
		}
		mod.sets.AddToUnion(fs.AllSetName, fs.ReadWriteSetName)
	}

	if mod.mem != nil {
		_, err := mod.sets.SetInfo(fs.MemorySetName)
		if err != nil {
			_, err = mod.sets.CreateSet(fs.MemorySetName, sets.TypeSet)
			if err != nil {
				return err
			}

			err = mod.sets.AddToUnion(fs.AllSetName, fs.MemorySetName)
			if err != nil {
				return err
			}
		}

		go events.Handle(context.Background(), &mod.events, func(ctx context.Context, event storage.EventDataCommitted) error {
			mod.sets.AddToSet(fs.MemorySetName, event.DataID)
			return nil
		})
	}

	// if we have file-based resources, use that as writable storage
	fileRes, ok := mod.assets.Res().(*resources.FileResources)
	if ok {
		dataPath := filepath.Join(fileRes.Root(), "data")
		err = os.MkdirAll(dataPath, 0700)
		if err == nil {
			err = mod.store.AddPath(dataPath)
			if err != nil {
				mod.log.Error("error adding writable data path: %v", err)
			}
		}
	} else {
		mod.mem = NewMemStore(&mod.events, 0)
	}

	for _, path := range mod.config.Store {
		mod.store.AddPath(path)
	}

	return nil
}
