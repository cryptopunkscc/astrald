package discovery

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "discovery"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:    node,
		config:  defaultConfig,
		log:     log.Tag(ModuleName),
		sources: map[Source]id.Identity{},
		cache:   map[string][]proto.ServiceEntry{},
	}

	mod.events.SetParent(node.Events())

	assets.LoadYAML(ModuleName, &mod.config)

	adm, err := modules.Find[*admin.Module](node.Modules())
	if err == nil {
		adm.AddCommand("discovery", NewAdmin(mod))
	}

	return mod, err
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
