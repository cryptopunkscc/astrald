package fs

import (
	"os"
	"path/filepath"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/resources"
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

	// set up the database
	mod.db = &DB{assets.Database()}

	err = mod.db.AutoMigrate(&dbLocalFile{})
	if err != nil {
		return nil, err
	}

	mod.fileIndexer, err = NewFileIndexer(mod.updateDbIndex, workers, updatesLen)
	if err != nil {
		return nil, err
	}

	mod.ops.AddStruct(mod, "Op")

	return mod, nil
}

func (mod *Module) addDefaultRepo() {
	res, ok := mod.assets.Res().(*resources.FileResources)
	if !ok {
		return
	}

	if _, ok := mod.repos.Get(DefaultRepoName); ok {
		return
	}

	// create the directory for the default repository if needed
	dataPath := filepath.Join(res.Root(), "data")

	err := os.MkdirAll(dataPath, 0700)
	if err != nil {
		return
	}

	repo := NewRepository(mod, "Default", dataPath)

	mod.Objects.AddRepository(DefaultRepoName, repo)
	mod.Objects.AddGroup(objects.RepoLocal, DefaultRepoName)
}

func init() {
	if err := core.RegisterModule(fs.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
