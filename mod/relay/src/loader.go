package relay

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/node"
)

type Loader struct{}

func (Loader) Load(node node.Node, assets assets.Assets, log *log.Logger) (node.Module, error) {
	var err error
	var mod = &Module{
		node:       node,
		log:        log.Tag(relay.ModuleName),
		assets:     assets,
		routes:     make(map[string]id.Identity),
		PathRouter: routers.NewPathRouter(node.Identity(), false),
	}

	_ = assets.LoadYAML(relay.ModuleName, &mod.config)

	mod.db = assets.Database()

	if err = mod.db.AutoMigrate(&dbCert{}); err != nil {
		return nil, err
	}

	return mod, err
}

func init() {
	if err := core.RegisterModule(relay.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
