package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

const taskQueueSize = 4096

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
		assets: assets,
		tasks:  make(chan func(ctx context.Context), taskQueueSize),
	}

	_ = assets.LoadYAML(shares.ModuleName, &mod.config)

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbRemoteShare{}, &dbRemoteData{}, &dbRemoteDesc{})
	if err != nil {
		return nil, err
	}

	err = mod.AddAuthorizer(&ACLAuthorizer{mod})
	if err != nil {
		return nil, err
	}

	err = mod.AddAuthorizer(&SelfAuthorizer{mod})
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(shares.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
