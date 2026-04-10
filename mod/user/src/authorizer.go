package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/user"
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

func (mod *Module) AuthorizeNodeRevokeContract(ctx *astral.Context, action *user.RevokeContractAction) bool {
	userID := mod.Identity()

	if action.Actor().IsEqual(userID) {
		return true
	}

	return false
}
