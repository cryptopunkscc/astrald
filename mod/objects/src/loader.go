package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
	}

	_ = assets.LoadYAML(objects.ModuleName, &mod.config)

	mod.ops.AddStruct(mod, "Op")

	mod.db = &DB{assets.Database()}

	err := mod.db.Migrate()
	if err != nil {
		return nil, err
	}

	mod.blueprints.Parent = astral.DefaultBlueprints
	mod.blueprints.Add(
		&objects.EventDiscovered{},
		&objects.EventCommitted{},
		&objects.SourcedObject{},
		&objects.SearchResult{},
	)

	return mod, nil
}

func init() {
	if err := core.RegisterModule(objects.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
