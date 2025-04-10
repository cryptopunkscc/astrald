package admin

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

var _ auth.Authorizer = &Module{}

func (mod *Module) Authorize(identity *astral.Identity, action auth.Action, target astral.Object) bool {
	switch action {
	case admin.ActionAccess:
		return mod.hasAccess(identity)
	case admin.ActionSudo:
		return identity.IsEqual(mod.node.Identity())
	}
	return false
}
