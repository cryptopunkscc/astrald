package content

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	var err error
	mod := &Module{
		node:  node,
		log:   log,
		ready: make(chan struct{}),
	}

	_ = assets.LoadYAML(content.ModuleName, &mod.config)

	mod.events.SetParent(node.Events())

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbDataType{}, &dbLabel{})
	if err != nil {
		return nil, err
	}

	mod.AddPrototypes(content.LabelDescriptor{}, content.TypeDescriptor{})

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(content.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
