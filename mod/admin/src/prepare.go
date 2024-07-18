package admin

import (
	"context"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/relay"
)

func (mod *Module) Prepare(ctx context.Context) (err error) {
	mod.relay, err = core.Load[relay.Module](mod.node, relay.ModuleName)
	if err != nil {
		return err
	}

	mod.keys, err = core.Load[keys.Module](mod.node, keys.ModuleName)
	if err != nil {
		return err
	}

	mod.auth, err = core.Load[auth.Module](mod.node, auth.ModuleName)
	if err != nil {
		return err
	}

	mod.dir, err = core.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	// load admins from config
	for _, name := range mod.config.Admins {
		adminID, err := mod.dir.Resolve(name)
		if err != nil {
			continue
		}

		mod.admins.Add(adminID.PublicKeyHex())
	}

	mod.auth.AddAuthorizer(mod)

	return nil
}
