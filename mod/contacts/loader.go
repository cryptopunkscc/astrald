package contacts

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "contacts"

type Loader struct{}

func (Loader) Load(node modules.Node, configStore config.Store) (modules.Module, error) {
	mod := &Module{
		node:   node,
		config: defaultConfig,
		log:    log.Tag("contacts"),
		ready:  make(chan struct{}),
	}

	mod.rootDir = configStore.BaseDir()

	configStore.LoadYAML(ModuleName, &mod.config)

	adm, err := modules.Find[*admin.Module](node.Modules())
	if err == nil {
		adm.AddCommand("contacts", NewAdmin(mod))
	}

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
