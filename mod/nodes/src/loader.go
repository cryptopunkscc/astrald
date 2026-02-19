package nodes

import (
	"time"

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
	mod.linkPool = NewLinkPool(mod, mod.peers)

	mod.RegisterNetworkStrategy("tcp", &BasicLinkStrategyFactory{mod: mod, network: "tcp"})
	mod.RegisterNetworkStrategy("tor", &TorLinkStrategyFactory{
		mod:     mod,
		network: "tor",
		config: TorLinkStrategyConfig{
			QuickRetries:      2,
			Retries:           3,
			SignalTimeout:     60 * time.Second,
			BackgroundTimeout: 360 * time.Second,
		},
	})
	mod.RegisterManualStrategy("nat", &NatLinkStrategyFactory{mod: mod})

	mod.db = &DB{assets.Database()}
	mod.dbResolver = &DBEndpointResolver{mod: mod}
	mod.resolvers.Add(mod.dbResolver)

	err = mod.db.AutoMigrate(&dbEndpoint{})
	if err != nil {
		return nil, err
	}

	mod.ops.AddStructPrefix(mod, "Op")

	return mod, err
}

func init() {
	if err := core.RegisterModule(nodes.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
