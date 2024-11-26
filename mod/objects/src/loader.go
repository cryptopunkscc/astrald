package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var mod = &Module{
		node:       node,
		config:     defaultConfig,
		log:        log,
		PathRouter: routers.NewPathRouter(node.Identity(), false),
	}

	mod.events.SetParent(node.Events())

	_ = assets.LoadYAML(objects.ModuleName, &mod.config)

	mod.db = assets.Database()

	// add core object prototypes
	var h astral.ObjectHeader
	var n astral.Nonce
	mod.addObjects(
		&h, &n,
		&astral.Identity{},
		&astral.Time{},
		&object.ID{},
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
