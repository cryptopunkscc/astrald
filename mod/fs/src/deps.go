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
	mod.storage.AddOpener(fs.ReadOnlySetName, mod.readonly, 30)

	// read write
	mod.storage.AddOpener(fs.ReadWriteSetName, mod.readwrite, 30)
	mod.storage.AddCreator(fs.ReadWriteSetName, mod.readwrite, 30)
	mod.storage.AddPurger(fs.ReadWriteSetName, mod.readwrite)

	// inject admin command
	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(fs.ModuleName, NewAdmin(mod))
	}

	mod.content.AddDescriber(mod)
	mod.content.AddPrototypes(fs.FileDesc{})
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
		mod.readonly.Add(filepath.Join(fileRes.Root(), "static_data"))

		dataPath := filepath.Join(fileRes.Root(), "data")
		err = os.MkdirAll(dataPath, 0700)
		if err == nil {
			err = mod.readwrite.AddPath(dataPath)
			if err != nil {
				mod.log.Error("error adding writable data path: %v", err)
			}
		}
	} else {
		// prefer memory for reads because of performance
		mod.storage.AddOpener(fs.MemorySetName, mod.memStore, 40)

		// avoid memory for writes because of its non-persistance
		mod.storage.AddCreator(fs.MemorySetName, mod.memStore, 0)

		mod.storage.AddPurger(fs.MemorySetName, mod.memStore)
	}

	for _, path := range mod.config.Store {
		mod.readwrite.AddPath(path)
	}

	return nil
}

func (mod *Module) createSets() error {
	var err error

	// Read-only
	mod.roSet, err = mod.sets.Open(fs.ReadOnlySetName, true)
	if err != nil {
		return err
	}

	// Read-write
	mod.rwSet, err = mod.sets.Open(fs.ReadWriteSetName, true)
	if err != nil {
		return err
	}

	// Memory
	mod.memSet, err = mod.sets.Open(fs.MemorySetName, true)
	if err != nil {
		return err
	}

	go events.Handle(context.Background(), &mod.memStore.events, func(event storage.EventDataCommitted) error {
		mod.memSet.Add(event.DataID)
		return nil
	})

	go events.Handle(context.Background(), mod.node.Events(), func(event storage.EventDataPurged) error {
		if _, found := mod.memStore.objects.Get(event.DataID.String()); found {
			return nil
		}
		mod.memSet.Remove(event.DataID)
		return nil
	})

	return nil
}
