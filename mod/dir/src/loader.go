package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/log"
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

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbAlias{})
	if err != nil {
		return nil, err
	}

	err = mod.setDefaultAlias()
	if err != nil {
		mod.log.Errorv(1, "error setting default alias: %v", err)
	}

	if cnode, ok := node.(*core.Node); ok {
		cnode.PushFormatFunc(func(v any) ([]log.Op, bool) {
			identity, ok := v.(id.Identity)
			if !ok {
				return nil, false
			}

			var color = log.Cyan

			if node.Identity().IsEqual(identity) {
				color = log.BrightGreen
			}

			var name = mod.DisplayName(identity)

			return []log.Op{
				log.OpColor{Color: color},
				log.OpText{Text: name},
				log.OpReset{},
			}, true
		})
	}

	return mod, nil
}

func init() {
	if err := core.RegisterModule(dir.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
