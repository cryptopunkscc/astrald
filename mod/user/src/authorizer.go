package user

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/node/authorizer"
)

var _ authorizer.Authorizer = &Authorizer{}

type Authorizer struct {
	mod *Module
}

func (auth *Authorizer) Authorize(identity id.Identity, action string, args ...any) bool {
	if identity.IsZero() {
		return false
	}

	if !identity.IsEqual(auth.mod.UserID()) {
		return false
	}

	switch action {
	case admin.AccessAction,
		objects.ActionRead,
		objects.ActionCreate,
		objects.ActionPurge,
		shares.DescribeAction,
		presence.ScanAction:
		return true
	}

	return false
}

func (auth *Authorizer) String() string {
	return "mod.user"
}
