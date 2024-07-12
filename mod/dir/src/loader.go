package dir

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/node"
)

type Loader struct{}

func (Loader) Load(node node.Node, assets assets.Assets, log *log.Logger) (node.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(dir.ModuleName, &mod.config)

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbAlias{})
	if err != nil {
		return nil, err
	}

	err = mod.node.Resolver().AddResolver(mod)
	if err != nil {
		return nil, err
	}

	err = mod.setDefaultAlias()
	if err != nil {
		mod.log.Errorv(1, "error setting default alias: %v", err)
	}

	return mod, nil
}

func init() {
	if err := core.RegisterModule(dir.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
