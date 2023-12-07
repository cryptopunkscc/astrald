package fs

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"os"
	"path/filepath"
)

const ModuleName = "fs"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    log,
		config: defaultConfig,
	}

	mod.events.SetParent(node.Events())

	_ = assets.LoadYAML(ModuleName, &mod.config)

	// set up database
	mod.db, err = assets.OpenDB(ModuleName)
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

	mod.storer = NewStorerService(mod)
	for _, path := range mod.config.Store {
		mod.storer.AddPath(path)
	}

	// if no storage paths are configured, use a default one
	if len(mod.config.Store) == 0 {
		if n, ok := node.(hasRootDir); ok {
			dataPath := filepath.Join(n.RootDir(), "data")
			err := os.MkdirAll(dataPath, 0600)
			if err == nil {
				mod.storer.AddPath(dataPath)
			}
		}
	}

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}

type hasRootDir interface {
	RootDir() string
}
