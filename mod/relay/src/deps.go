package relay

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.content, err = modules.Load[content.Module](mod.node, content.ModuleName)
	if err != nil {
		return err
	}

	mod.storage, err = modules.Load[storage.Module](mod.node, storage.ModuleName)
	if err != nil {
		return err
	}

	mod.keys, err = modules.Load[keys.Module](mod.node, keys.ModuleName)
	if err != nil {
		return err
	}
	_ = mod.content.AddDescriber(mod)

	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(relay.ModuleName, NewAdmin(mod))
	}

	mod.node.Router().AddRoute(id.Anyone, id.Anyone, mod, 20)

	mod.content.AddPrototypes(relay.CertDescriptor{})

	return nil
}
