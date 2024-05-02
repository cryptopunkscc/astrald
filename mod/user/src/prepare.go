package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) Prepare(ctx context.Context) error {
	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(user.ModuleName, NewAdmin(mod))
	}

	if local := mod.config.Identity; local != "" {
		userID, err := mod.node.Resolver().Resolve(local)
		if err != nil {
			mod.log.Error("config: cannot resolve local user %v: %v", local, err)
		}

		err = mod.SetUserID(userID)
		if err != nil {
			mod.log.Error("SetLocalUser: %v", err)
		}
	} else {
		userID, _ := mod.loadUserID()
		if !userID.IsZero() {
			mod.setUserID(userID)
		}
	}

	// look for user profiles in discovered services
	go mod.discoverUsers(ctx)

	return nil
}
