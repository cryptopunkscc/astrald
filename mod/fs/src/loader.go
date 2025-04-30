package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/resources"
	"os"
	"path/filepath"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		assets: assets,
		log:    log,
		config: defaultConfig,
	}

	_ = assets.LoadYAML(fs.ModuleName, &mod.config)

	// set up database
	mod.db = &DB{assets.Database()}

	err = mod.db.AutoMigrate(&dbLocalFile{})
	if err != nil {
		return nil, err
	}

	for name, path := range mod.config.Repos {
		mod.repos.Set(name, NewRepository(mod, name, path))
	}

	for name, path := range mod.config.Watch {
		repo, err := NewWatchRepository(mod, path, name)
		if err != nil {
			mod.log.Error("error adding watch repo %v: %v", name, err)
			continue
		}
		mod.repos.Set(name, repo)
		mod.log.Info("watching %v as %v", path, name)
	}

	res, ok := mod.assets.Res().(*resources.FileResources)
	if ok {
		// create default repository if needed
		if _, ok := mod.repos.Get("default"); !ok {
			dataPath := filepath.Join(res.Root(), "data")
			if os.MkdirAll(dataPath, 0700) == nil {
				mod.repos.Set("default", NewRepository(mod, "default", dataPath))
			}
		}
	}

	mod.ops.AddStruct(mod, "Op")

	return mod, nil
}

func init() {
	if err := core.RegisterModule(fs.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
