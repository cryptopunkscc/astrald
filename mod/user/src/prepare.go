package user

import (
	"context"
)

func (mod *Module) Prepare(ctx context.Context) error {
	if local := mod.config.Identity; local != "" {
		userID, err := mod.Dir.Resolve(local)
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
			mod.SetUserID(userID)
		}
	}

	return nil
}
