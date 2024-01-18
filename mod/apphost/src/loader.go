package apphost

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"net"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	var err error

	mod := &Module{
		config:    defaultConfig,
		node:      node,
		listeners: make([]net.Listener, 0),
		guests:    make(map[string]*Guest),
		execs:     []*Exec{},
		log:       log,
	}

	_ = assets.LoadYAML(apphost.ModuleName, &mod.config)

	// set up database
	mod.db, err = assets.OpenDB(apphost.ModuleName)
	if err != nil {
		return nil, err
	}

	err = mod.db.AutoMigrate(&dbAccessToken{})
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(apphost.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
