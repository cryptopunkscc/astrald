package archives

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
	}

	mod.events.SetParent(node.Events())

	_ = assets.LoadYAML(archives.ModuleName, &mod.config)

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbArchive{}, &dbEntry{})
	if err != nil {
		return nil, err
	}

	node.Auth().Add(&Authorizer{mod: mod})

	return mod, err
}

func init() {
	if err := modules.RegisterModule(archives.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
