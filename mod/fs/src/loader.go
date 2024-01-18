package fs

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/resources"
	"os"
	"path/filepath"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    log,
		config: defaultConfig,
	}

	mod.events.SetParent(node.Events())

	_ = assets.LoadYAML(fs.ModuleName, &mod.config)

	// set up database
	mod.db, err = assets.OpenDB(fs.ModuleName)
	if err != nil {
		return nil, err
	}

	if err := mod.db.AutoMigrate(&dbLocalFile{}); err != nil {
		return nil, err
	}

	// set up services
	mod.indexer = NewIndexerService(mod)
	for _, path := range mod.config.Index {
		mod.indexer.Add(path)
	}

	mod.store = NewStoreService(mod)
	for _, path := range mod.config.Store {
		mod.store.AddPath(path)
	}

	// if we have file-based resources, use that as writable storage
	fileRes, ok := assets.Res().(*resources.FileResources)
	if ok {
		dataPath := filepath.Join(fileRes.Root(), "data")
		err = os.MkdirAll(dataPath, 0700)
		if err == nil {
			err = mod.store.AddPath(dataPath)
			if err != nil {
				mod.log.Error("error adding writable data path: %v", err)
			}
		}
	}

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(fs.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
