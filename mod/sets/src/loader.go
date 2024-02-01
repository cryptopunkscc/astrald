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

	err = mod.db.AutoMigrate(&dbSet{}, &dbMember{}, &dbSetInclusion{})
	if err != nil {
		return nil, err
	}

	mod.SetOpener(sets.TypeBasic, func(name string) (sets.Set, error) {
		return mod.basicOpener(name)
	})
	mod.SetOpener(sets.TypeUnion, func(name string) (sets.Set, error) {
		return mod.unionOpener(name)
	})

	mod.universe, err = sets.Open[*UnionSet](mod, sets.UniverseSet)
	if err != nil {
		mod.universe, err = mod.createUnion(sets.UniverseSet)
		if err != nil {
			return nil, err
		}
		mod.SetVisible(sets.UniverseSet, true)
		mod.SetDescription(sets.UniverseSet, "All data from everywhere")
	}

	mod.localnode, err = sets.Open[*UnionSet](mod, sets.LocalNodeSet)
	if err != nil {
		mod.localnode, err = mod.createUnion(sets.LocalNodeSet)
		if err != nil {
			return nil, err
		}
		mod.universe.Add(sets.LocalNodeSet)
		mod.SetVisible(sets.LocalNodeSet, true)
		mod.SetDescription(sets.LocalNodeSet, "All data on this node")
	}

	return mod, err
}

func init() {
	if err := modules.RegisterModule(sets.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
