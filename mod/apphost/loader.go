package apphost

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"net"
)

const ModuleName = "apphost"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
	var err error

	mod := &Module{
		config:    defaultConfig,
		node:      node,
		listeners: make([]net.Listener, 0),
		tokens:    make(map[string]id.Identity),
		guests:    make(map[string]*Guest),
		execs:     []*Exec{},
		log:       log,
	}

	_ = assets.LoadYAML(ModuleName, &mod.config)

	mod.keys, err = assets.KeyStore()
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
