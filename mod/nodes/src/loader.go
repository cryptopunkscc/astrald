package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    log,
		assets: assets,
		in:     make(chan *Frame),
	}

	_ = assets.LoadYAML(nodes.ModuleName, &mod.config)

	mod.peers = NewPeers(mod)

	mod.db = &DB{assets.Database()}
	mod.dbResolver = &DBEndpointResolver{mod: mod}
	mod.resolvers.Add(mod.dbResolver)

	err = mod.db.AutoMigrate(&dbEndpoint{}, &dbService{})
	if err != nil {
		return nil, err
	}

	mod.db.DeleteExpiredServices()

	mod.ops.AddStruct(mod, "Op")

	return mod, err
}

func init() {
	if err := core.RegisterModule(nodes.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
