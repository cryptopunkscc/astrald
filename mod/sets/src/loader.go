package sets

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/sets"
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

	_ = assets.LoadYAML(sets.ModuleName, &mod.config)

	mod.events.SetParent(node.Events())

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbSet{}, &dbMember{})
	if err != nil {
		return nil, err
	}

	return mod, err
}

func init() {
	if err := core.RegisterModule(sets.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
