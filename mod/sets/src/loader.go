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
		config: defaultConfig,
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

	// basic set has no additional wrapper
	mod.SetWrapper(sets.TypeBasic, func(s sets.Set) (sets.Set, error) {
		return s, nil
	})

	// add a wrapper for a union set
	mod.SetWrapper(sets.TypeUnion, mod.unionWrapper)

	// create core sets
	mod.device, err = mod.OpenOrCreateUnion(sets.DeviceSet)
	if err != nil {
		return nil, err
	}
	mod.device.SetDisplayName(mod.config.Display.Device)

	mod.virtual, err = mod.OpenOrCreateUnion(sets.VirtualSet)
	if err != nil {
		return nil, err
	}
	mod.virtual.SetDisplayName(mod.config.Display.Virtual)

	mod.network, err = mod.OpenOrCreateUnion(sets.NetworkSet)
	if err != nil {
		return nil, err
	}
	mod.network.SetDisplayName(mod.config.Display.Network)

	mod.universe, err = mod.OpenOrCreateUnion(sets.UniverseSet)
	if err != nil {
		return nil, err
	}
	mod.universe.SetDisplayName(mod.config.Display.Universe)

	mod.universe.AddSubset(sets.DeviceSet, sets.VirtualSet, sets.NetworkSet)

	return mod, err
}

func init() {
	if err := modules.RegisterModule(sets.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
