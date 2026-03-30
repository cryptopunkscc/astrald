package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/user"
)

func (mod *Module) AuthorizeUser(_ *astral.Context, identity *astral.Identity, target astral.Object) bool {
	return identity.IsEqual(mod.Identity())
}

func (mod *Module) AuthorizeUserRevokeContract(_ *astral.Context, identity *astral.Identity, contract *user.SignedNodeContract) bool {
	return identity.IsEqual(mod.Identity())
}

func (mod *Module) AuthorizeNodeRead(ctx *astral.Context, identity *astral.Identity, target astral.Object) bool {
	for _, u := range mod.ActiveUsers(identity) {
		if mod.Auth.Authorize(ctx, u, objects.ActionRead, target) {
			return true
		}
	}
	return false
}

func (mod *Module) AuthorizeNodeCreate(ctx *astral.Context, identity *astral.Identity, target astral.Object) bool {
	for _, u := range mod.ActiveUsers(identity) {
		if mod.Auth.Authorize(ctx, u, objects.ActionCreate, target) {
			return true
		}
	}
	return false
}

func (mod *Module) AuthorizeNodeRelay(_ *astral.Context, identity *astral.Identity, targetID *astral.Identity) bool {
	for _, u := range mod.ActiveUsers(identity) {
		if u.IsEqual(targetID) {
			return true
		}
	}
	return false
}

func (mod *Module) AuthorizeNodeRevokeContract(_ *astral.Context, identity *astral.Identity, contract *user.SignedNodeContract) bool {
	userID := mod.Identity()

	if contract.UserID.IsEqual(userID) {
		return true
	}

	for _, u := range mod.ActiveUsers(identity) {
		if u.IsEqual(userID) {
			return true
		}
	}

	return false
}
