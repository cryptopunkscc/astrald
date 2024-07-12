package apphost

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/node"
	"net"
)

type Loader struct{}

func (Loader) Load(node node.Node, assets assets.Assets, log *log.Logger) (node.Module, error) {
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
	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbAccessToken{})
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
