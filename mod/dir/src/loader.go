package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/astral/term"
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

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbAlias{})
	if err != nil {
		return nil, err
	}

	err = mod.setDefaultAlias()
	if err != nil {
		mod.log.Errorv(1, "error setting default alias: %v", err)
	}

	term.SetTranslateFunc(func(identity *astral.Identity) astral.Object {
		var color = "cyan"

		if node.Identity().IsEqual(identity) {
			color = "brightgreen"
		}

		var name = mod.DisplayName(identity)

		return &term.ColorString{
			Color: astral.String8(color),
			Text:  astral.String32(name),
		}
	})

	mod.resolvers.Add(&DNS{Module: mod})

	return mod, nil
}

func init() {
	if err := core.RegisterModule(dir.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
