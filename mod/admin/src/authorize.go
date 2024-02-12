package admin

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/node/authorizer"
)

var _ authorizer.Authorizer = &Module{}

func (mod *Module) Authorize(identity id.Identity, action string, args ...any) bool {
	if action != admin.AccessAction {
		return false
	}

	return mod.hasAccess(identity)
}
