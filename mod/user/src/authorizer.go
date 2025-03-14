package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/status"
)

func (mod *Module) Authorize(identity *astral.Identity, action auth.Action, target astral.Object) bool {
	if identity.IsZero() {
		return false
	}

	// allow the user to perform whitelisted actions
	if identity.IsEqual(mod.UserID()) {
		switch action {
		case admin.ActionAccess,
			admin.ActionSudo,
			fs.ActionManage,
			objects.ActionRead,
			objects.ActionWrite,
			objects.ActionPurge,
			objects.ActionSearch,
			objects.ActionReadDescriptor,
			status.ActionList:
			return true
		}
	}

	// nodes inherit some permissions from their owner
	if owner := mod.Owner(identity); !owner.IsZero() {
		switch action {
		case objects.ActionRead,
			objects.ActionWrite,
			objects.ActionPurge,
			objects.ActionSearch,
			objects.ActionReadDescriptor:
			if mod.Authorize(owner, action, target) {
				return true
			}

		case nodes.ActionRelayFor:
			t, ok := target.(*astral.Identity)
			if !ok {
				break
			}

			if _, err := mod.findContractID(t, identity); err == nil {
				return true
			}

		}

	}

	return false
}
