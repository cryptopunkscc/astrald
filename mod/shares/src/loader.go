package shares

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(shares.ModuleName, &mod.config)

	mod.db, err = assets.OpenDB(shares.ModuleName)
	if err != nil {
		return nil, err
	}

	err = mod.db.AutoMigrate(&dbRemoteShare{}, &dbRemoteData{})
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
