package fs

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/sig"
)

type Loader struct{}

func (Loader) Load(node node.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:    node,
		assets:  assets,
		log:     log,
		config:  defaultConfig,
		updates: make(chan sig.Task, updatesLen),
	}

	mod.events.SetParent(node.Events())

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

	for _, path := range mod.config.Watch {
		mod.Watch(path)
	}

	for _, path := range mod.config.Store {
		mod.Watch(path)
	}

	return mod, nil
}

func init() {
	if err := core.RegisterModule(fs.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
