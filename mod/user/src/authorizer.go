package user

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/object"
)

var _ node.Authorizer = &Authorizer{}

type Authorizer struct {
	mod *Module
}

func (auth *Authorizer) Authorize(identity id.Identity, action string, args ...any) bool {
	if identity.IsZero() {
		return false
	}

	// make our node contracts publicly available
	if action == objects.ActionRead {
		if len(args) > 0 {
			if objectID, ok := args[0].(object.ID); ok {
				cache := auth.mod.getCache(objectID)
				if cache != nil {
					if cache.UserID.IsEqual(auth.mod.UserID()) {
						return true
					}
				}
			}
		}
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
			shares.DescribeAction,
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
			if auth.Authorize(owner, action, args...) {
				return true
			}
		}

	}

	return false
}

func (auth *Authorizer) String() string {
	return "mod." + user.ModuleName
}
