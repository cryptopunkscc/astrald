package fwd

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
)

const ModuleName = "fwd"

type Loader struct{}

func (Loader) Load(node node.Node, assets assets.Assets, log *log.Logger) (node.Module, error) {
	mod := &Module{
		node:    node,
		config:  defaultConfig,
		log:     log,
		servers: make(map[*ServerRunner]struct{}),
	}

	_ = assets.LoadYAML(ModuleName, &mod.config)

	return mod, nil
}

func init() {
	if err := core.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
