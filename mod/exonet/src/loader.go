package exonet

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/node"
)

type Loader struct{}

func (Loader) Load(node node.Node, assets assets.Assets, log *log.Logger) (node.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(exonet.ModuleName, &mod.config)

	return mod, err
}

func init() {
	if err := core.RegisterModule(exonet.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
