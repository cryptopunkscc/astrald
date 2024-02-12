package relay

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    log.Tag(relay.ModuleName),
		assets: assets,
		routes: make(map[string]id.Identity),
	}

	_ = assets.LoadYAML(relay.ModuleName, &mod.config)

	mod.db = assets.Database()

	if err = mod.db.AutoMigrate(&dbCert{}); err != nil {
		return nil, err
	}

	return mod, err
}

func init() {
	if err := modules.RegisterModule(relay.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
