package fs

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"os"
	"path/filepath"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
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
	mod.index = NewIndexService(mod)
	for _, path := range mod.config.Index {
		mod.index.Add(path)
	}

	mod.store = NewStoreService(mod)
	for _, path := range mod.config.Store {
		mod.store.AddPath(path)
	}

	// if no storage paths are configured, use a default one
	if len(mod.config.Store) == 0 {
		if n, ok := node.(hasRootDir); ok {
			dataPath := filepath.Join(n.RootDir(), "data")
			err := os.MkdirAll(dataPath, 0700)
			if err == nil {
				mod.store.AddPath(dataPath)
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

type hasRootDir interface {
	RootDir() string
}
