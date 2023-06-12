package contacts

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "contacts"
const DatabaseName = "contacts.db"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
	}

	assets.LoadYAML(ModuleName, &mod.config)

	mod.db, err = assets.OpenDB(DatabaseName)
	if err != nil {
		return nil, err
	}

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
