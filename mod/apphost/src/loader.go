package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"net"
)

const (
	mRegisterApp    = "apps.register_app"
	mUnregisterApp  = "apps.unregister_app"
	mGetAccessToken = "apps.get_access_token"
	mList           = "apps.list"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error

	mod := &Module{
		config:    defaultConfig,
		node:      node,
		listeners: make([]net.Listener, 0),
		guests:    make(map[string]*Guest),
		execs:     []*Exec{},
		log:       log,
		router:    routers.NewPathRouter(node.Identity(), false),
	}

	_ = assets.LoadYAML(apphost.ModuleName, &mod.config)

	// set up database
	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbAccessToken{}, &dbApp{})
	if err != nil {
		return nil, err
	}

	mod.router.AddRouteFunc(mRegisterApp, mod.regsiterApp)
	mod.router.AddRouteFunc(mUnregisterApp, mod.unregsiterApp)
	mod.router.AddRouteFunc(mGetAccessToken, mod.getAccessToken)
	mod.router.AddRouteFunc(mList, mod.list)

	return mod, nil
}

func init() {
	if err := core.RegisterModule(apphost.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
