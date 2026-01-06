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
		node:           node,
		log:            l,
		serviceEnabled: true,
	}

	mod.pool = NewPairPool(mod)
	mod.ops.AddStruct(mod, "Op")

	// Initialize pure service feed (broadcast-only).
	mod.serviceFeed = services.NewServiceFeed()

	return mod, nil
}

func init() {
	err := core.RegisterModule(nat.ModuleName, Loader{})
	if err != nil {
		panic(err)
	}
}
