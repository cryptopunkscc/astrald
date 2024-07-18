package content

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
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

	return mod, nil
}

func init() {
	if err := core.RegisterModule(content.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
