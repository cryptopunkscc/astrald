package ether

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/ether"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(ether.ModuleName, &mod.config)

	if err = mod.setupSocket(); err != nil {
		return nil, err
	}

	return mod, err
}

func init() {
	if err := core.RegisterModule(ether.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
