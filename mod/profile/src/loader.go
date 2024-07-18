package profile

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
)

const ModuleName = "profile"

type Loader struct{}

func (Loader) Load(node astral.Node, _ assets.Assets, log *log.Logger) (core.Module, error) {
	mod := &Module{
		node:       node,
		log:        log,
		PathRouter: routers.NewPathRouter(node.Identity(), false),
	}

	return mod, nil
}

func init() {
	if err := core.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
