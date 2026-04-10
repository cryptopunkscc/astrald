package user

import (
	"github.com/cryptopunkscc/astrald/astral"
)

func (mod *Module) AuthorizeUser(_ *astral.Context, identity *astral.Identity, target astral.Object) bool {
	return identity.IsEqual(mod.Identity())
}

func (mod *Module) AuthorizeNodeRelay(_ *astral.Context, identity *astral.Identity, targetID *astral.Identity) bool {
	for _, u := range mod.ActiveUsers(identity) {
		if u.IsEqual(targetID) {
			return true
		}
	}
	return false
}
