package kos

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/kos"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
		assets: assets,
		db:     &DB{assets.Database()},
	}

	_ = assets.LoadYAML(kos.ModuleName, &mod.config)

	err = mod.db.AutoMigrate(&dbEntry{})

	mod.ops.AddStruct(mod, "Op")
	
	return mod, err
}

func init() {
	if err := core.RegisterModule(kos.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
