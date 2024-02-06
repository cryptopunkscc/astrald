package webdata

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/webdata"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:     node,
		log:      log,
		config:   defaultConfig,
		assets:   assets,
		identity: node.Identity(),
	}

	_ = assets.LoadYAML(webdata.ModuleName, &mod.config)

	if mod.config.Identity != "" {
		mod.identity, err = mod.node.Resolver().Resolve(mod.config.Identity)
		if err != nil {
			return nil, fmt.Errorf("config: cannot resolve identity: %w", err)
		}
	}

	return mod, err
}

func init() {
	if err := modules.RegisterModule(webdata.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
