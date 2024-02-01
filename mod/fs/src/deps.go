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

	// inject admin command
	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(fs.ModuleName, NewAdmin(mod))
	}

	if mod.content != nil {
		mod.content.AddDescriber(mod)
	}

	mod.memStore = NewMemStore(&mod.events, 0)

	// wait for data module to finish preparing
	ctx, cancel := context.WithTimeoutCause(context.Background(), 15*time.Second, errors.New("data module timed out"))
	defer cancel()
	if err := mod.content.Ready(ctx); err != nil {
		return err
	}

	err = mod.createSets()
	if err != nil {
		return err
	}

	// if we have file-based resources, use that as writable storage
	fileRes, ok := mod.assets.Res().(*resources.FileResources)
	if ok {
		mod.indexer.Add(filepath.Join(fileRes.Root(), "static_data"))

		dataPath := filepath.Join(fileRes.Root(), "data")
		err = os.MkdirAll(dataPath, 0700)
		if err == nil {
			err = mod.store.AddPath(dataPath)
			if err != nil {
				mod.log.Error("error adding writable data path: %v", err)
			}
		}
	} else {
		mod.storage.AddReader(fs.MemorySetName, mod.memStore)
		mod.storage.AddStore(fs.MemorySetName, mod.memStore)
	}

	for _, path := range mod.config.Store {
		mod.store.AddPath(path)
	}

	return nil
}

func (mod *Module) createSets() error {
	var err error

	// All
	mod.allSet, err = sets.Open[sets.Union](mod.sets, fs.AllSetName)
	if err != nil {
		mod.allSet, err = mod.sets.CreateUnion(fs.AllSetName)
		if err != nil {
			return err
		}
		mod.sets.Localnode().Add(fs.AllSetName)
		mod.sets.SetVisible(fs.AllSetName, true)
		mod.sets.SetDescription(fs.AllSetName, "Local filesystem")
	}

	// Read-only
	mod.roSet, err = sets.Open[sets.Basic](mod.sets, fs.ReadOnlySetName)
	if err != nil {
		mod.roSet, err = mod.sets.CreateBasic(fs.ReadOnlySetName)
		if err != nil {
			return err
		}
		mod.allSet.Add(fs.ReadOnlySetName)
	}

	// Read-write
	mod.rwSet, err = sets.Open[sets.Basic](mod.sets, fs.ReadWriteSetName)
	if err != nil {
		mod.rwSet, err = mod.sets.CreateBasic(fs.ReadWriteSetName)
		if err != nil {
			return err
		}
		mod.allSet.Add(fs.ReadWriteSetName)
	}

	// Memory set
	if mod.memStore != nil {
		mod.memSet, err = sets.Open[sets.Basic](mod.sets, fs.MemorySetName)
		if err != nil {
			mod.memSet, err = mod.sets.CreateBasic(fs.MemorySetName)
			if err != nil {
				return err
			}
			mod.allSet.Add(fs.MemorySetName)
		}

		go events.Handle(context.Background(), &mod.memStore.events, func(ctx context.Context, event storage.EventDataCommitted) error {
			mod.memSet.Add(event.DataID)
			return nil
		})
	}

	return nil
}
