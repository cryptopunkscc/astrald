package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/mod/user"
)

var _ auth.Authorizer = &Authorizer{}

type Authorizer struct {
	mod *Module
}

func (auth *Authorizer) Authorize(identity id.Identity, action string, target astral.Object) bool {
	if identity.IsZero() {
		return false
	}

	// allow the user to perform whitelisted actions
	if identity.IsEqual(auth.mod.UserID()) {
		switch action {
		case admin.ActionAccess,
			admin.ActionSudo,
			objects.ActionRead,
			objects.ActionWrite,
			objects.ActionPurge,
			objects.ActionSearch,
			presence.ScanAction:
			return true
		}
	}

	// nodes inherit some permissions from their owner
	if owner := auth.mod.Owner(identity); !owner.IsZero() {
		switch action {
		case objects.ActionRead,
			objects.ActionWrite,
			objects.ActionPurge,
			objects.ActionSearch:
			if auth.Authorize(owner, action, target) {
				return true
			}
		}

	}

	return false
}

func (auth *Authorizer) String() string {
	return "mod." + user.ModuleName
}
