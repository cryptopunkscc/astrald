package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/node"
)

type Loader struct{}

const taskQueueSize = 4096

func (Loader) Load(node node.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:       node,
		config:     defaultConfig,
		log:        log,
		assets:     assets,
		tasks:      make(chan func(ctx context.Context), taskQueueSize),
		PathRouter: routers.NewPathRouter(node.Identity(), false),
	}

	_ = assets.LoadYAML(shares.ModuleName, &mod.config)

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbRemoteShare{}, &dbRemoteData{}, &dbRemoteDesc{})
	if err != nil {
		return nil, err
	}

	err = mod.addAuthorizer(&ACLAuthorizer{mod})
	if err != nil {
		return nil, err
	}

	mod.node.Auth().Add(&Authorizer{mod: mod})

	return mod, nil
}

func init() {
	if err := core.RegisterModule(shares.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
