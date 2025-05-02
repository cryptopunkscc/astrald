package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) Authorize(identity *astral.Identity, action auth.Action, target astral.Object) bool {
	if identity.IsZero() {
		return false
	}

	var contract = mod.ActiveContract()

	// allow the user to perform whitelisted actions
	if contract != nil && identity.IsEqual(contract.UserID) {
		switch action {
		case admin.ActionAccess,
			admin.ActionSudo,
			objects.ActionRead,
			objects.ActionCreate:
			return true
		}
	}

	// nodes inherit some permissions from their users
	for _, userID := range mod.ActiveUsers(identity) {
		switch action {
		case objects.ActionRead,
			objects.ActionCreate:
			if mod.Authorize(userID, action, target) {
				return true
			}

		case nodes.ActionRelayFor:
			userID, ok := target.(*astral.Identity)
			if !ok {
				break
			}

			for _, u := range mod.ActiveUsers(identity) {
				if u.IsEqual(userID) {
					return true
				}
			}
		}
	}

	return false
}
