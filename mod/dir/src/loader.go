package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/dir"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, l *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    l,
		assets: assets,
	}

	_ = assets.LoadYAML(dir.ModuleName, &mod.config)

	mod.ops.AddStruct(mod, "Op")

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbAlias{})
	if err != nil {
		return nil, err
	}

	err = mod.setDefaultAlias()
	if err != nil {
		mod.log.Errorv(1, "error setting default alias: %v", err)
	}

	mod.resolvers.Add(&DNS{Module: mod})

	return mod, nil
}

func init() {
	if err := core.RegisterModule(dir.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
