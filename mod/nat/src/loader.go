package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/services"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, l *log.Logger) (core.Module, error) {
	mod := &Module{
		node: node,
		log:  l,
	}

	mod.pool = NewPairPool(mod)
	mod.ops.AddStruct(mod, "Op")

	// Initialize service feed with initial NAT service state
	mod.serviceFeed = services.NewServiceFeed(&services.Service{
		Name:        nat.ModuleName,
		Identity:    node.Identity(),
		Composition: astral.NewBundle(),
	})

	return mod, nil
}

func init() {
	err := core.RegisterModule(nat.ModuleName, Loader{})
	if err != nil {
		panic(err)
	}
}
