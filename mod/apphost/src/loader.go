package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"net"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error

	mod := &Module{
		config:    defaultConfig,
		node:      node,
		listeners: make([]net.Listener, 0),
		log:       log,
	}

	_ = assets.LoadYAML(apphost.ModuleName, &mod.config)

	mod.scope.AddStruct(mod, "Op")

	// set up the database
	mod.db = &DB{assets.Database()}

	err = mod.db.AutoMigrate(&dbAccessToken{}, &dbAppContract{})
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func init() {
	if err := core.RegisterModule(apphost.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
