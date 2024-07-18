package setup

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/setup"
	"github.com/cryptopunkscc/astrald/mod/user"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.user, err = core.Load[user.Module](mod.node, user.ModuleName)
	if err != nil {
		return err
	}

	mod.keys, err = core.Load[keys.Module](mod.node, keys.ModuleName)
	if err != nil {
		return err
	}

	mod.apphost, err = core.Load[apphost.Module](mod.node, apphost.ModuleName)
	if err != nil {
		return err
	}

	mod.presence, err = core.Load[presence.Module](mod.node, presence.ModuleName)
	if err != nil {
		return err
	}

	mod.relay, err = core.Load[relay.Module](mod.node, relay.ModuleName)
	if err != nil {
		return err
	}

	mod.dir, err = core.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	mod.presence.AddHookAdOut(mod)

	if adm, err := core.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(setup.ModuleName, NewAdmin(mod))
	}

	return nil
}
