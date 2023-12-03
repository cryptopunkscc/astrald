package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	_assets "github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"path/filepath"
)

const ModuleName = "storage"
const configName = "storage"

type Loader struct{}

func (Loader) Load(node modules.Node, assets _assets.Store, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:           node,
		config:         defaultConfig,
		readers:        make([]storage.Reader, 0),
		accessCheckers: make(map[AccessChecker]struct{}, 0),
		log:            log,
	}

	mod.events.SetParent(node.Events())

	_ = assets.LoadYAML(configName, &mod.config)

	mod.db, err = assets.OpenDB(ModuleName)
	if err != nil {
		return nil, err
	}

	if err = mod.setupDatabase(); err != nil {
		return nil, err
	}

	mod.localFiles = NewLocalFiles(mod)
	mod.AddReader(mod.localFiles)
	for _, path := range mod.config.LocalFiles {
		mod.localFiles.AddDir(context.Background(), path)
	}

	if fs, ok := node.(hasRootDir); ok {
		var root = fs.RootDir()
		if root != "" {
			dataDir := filepath.Join(root, "data")
			mod.localFiles.AddDir(context.Background(), dataDir)
		}
	}

	return mod, nil
}

type hasRootDir interface {
	RootDir() string
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
