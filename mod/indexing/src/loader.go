package indexing

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/indexing"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(indexing.ModuleName, &mod.config)

	mod.ops.AddStruct(mod, "Op")

	mod.db, err = newDB(assets.Database())
	if err != nil {
		return nil, err
	}

	err = mod.db.autoMigrate()
	if err != nil {
		return nil, err
	}

	return mod, err
}

func init() {
	if err := core.RegisterModule(indexing.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
