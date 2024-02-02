package zip

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/zip"
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

	_ = assets.LoadYAML(zip.ModuleName, &mod.config)

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbZip{}, &dbContents{})
	if err != nil {
		return nil, err
	}

	return mod, err
}

func init() {
	if err := modules.RegisterModule(zip.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
