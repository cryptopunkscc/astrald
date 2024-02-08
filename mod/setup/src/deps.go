package setup

import (
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/setup"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.user, err = modules.Load[user.Module](mod.node, user.ModuleName)
	if err != nil {
		return err
	}

	mod.keys, err = modules.Load[keys.Module](mod.node, keys.ModuleName)
	if err != nil {
		return err
	}

	mod.apphost, err = modules.Load[apphost.Module](mod.node, apphost.ModuleName)
	if err != nil {
		return err
	}

	mod.presence, err = modules.Load[presence.Module](mod.node, presence.ModuleName)
	if err != nil {
		return err
	}

	mod.relay, err = modules.Load[relay.Module](mod.node, relay.ModuleName)
	if err != nil {
		return err
	}

	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(setup.ModuleName, NewAdmin(mod))
	}

	return nil
}
