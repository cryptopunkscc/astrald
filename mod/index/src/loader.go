package index

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/index"
	"github.com/cryptopunkscc/astrald/mod/wallet"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(wallet.ModuleName, &mod.config)

	mod.events.SetParent(node.Events())

	mod.db, err = assets.OpenDB(index.ModuleName)
	if err != nil {
		return nil, err
	}

	err = mod.db.AutoMigrate(&dbIndex{}, &dbEntry{})
	if err != nil {
		return nil, err
	}

	return mod, err
}

func init() {
	if err := modules.RegisterModule(wallet.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
