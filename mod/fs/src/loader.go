package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
	"os"
	"path/filepath"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:    node,
		assets:  assets,
		log:     log,
		config:  defaultConfig,
		updates: make(chan sig.Task, updatesLen),
	}

	_ = assets.LoadYAML(fs.ModuleName, &mod.config)

	// set up database
	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbLocalFile{})
	if err != nil {
		return nil, err
	}

	mod.watcher, err = NewWatcher()
	if err != nil {
		return nil, err
	}

	mod.watcher.OnWriteDone = mod.onWriteDone
	mod.watcher.OnRemoved = mod.enqueueUpdate
	mod.watcher.OnRenamed = mod.enqueueUpdate
	mod.watcher.OnChmod = mod.enqueueUpdate
	mod.watcher.OnDirCreated = func(s string) {
		mod.watcher.Add(s, true)
	}

	for name, path := range mod.config.Repos {
		mod.repos.Set(name, NewRepository(mod, name, path))
	}

	res, ok := mod.assets.Res().(*resources.FileResources)
	if ok {
		mod.Watch(filepath.Join(res.Root(), "static_data"))

		// create default repository if neeeded
		if _, ok := mod.repos.Get("default"); !ok {
			dataPath := filepath.Join(res.Root(), "data")
			if os.MkdirAll(dataPath, 0700) == nil {
				mod.repos.Set("default", NewRepository(mod, "default", dataPath))
				mod.Watch(dataPath)
			}
		}
	}

	mod.ops.AddOp("info", mod.opInfo)

	return mod, nil
}

func init() {
	if err := core.RegisterModule(fs.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
