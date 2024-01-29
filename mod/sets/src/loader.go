package sets

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(sets.ModuleName, &mod.config)

	mod.events.SetParent(node.Events())

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbSet{}, &dbMember{}, &dbUnion{})
	if err != nil {
		return nil, err
	}

	if _, err := mod.SetInfo(sets.LocalNodeSet); err != nil {
		_, err = mod.CreateSet(sets.LocalNodeSet, sets.TypeUnion)
		if err != nil {
			return nil, err
		}
	}

	return mod, err
}

func init() {
	if err := modules.RegisterModule(sets.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
