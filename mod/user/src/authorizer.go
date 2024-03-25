package user

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/authorizer"
)

var _ authorizer.Authorizer = &Authorizer{}

type Authorizer struct {
	mod *Module
}

func (auth *Authorizer) Authorize(identity id.Identity, action string, args ...any) bool {
	var localUser = auth.mod.LocalUser()
	if localUser == nil {
		return false
	}
	var userID = localUser.Identity()
	if userID.IsZero() {
		return false
	}
	if !userID.IsEqual(identity) {
		return false
	}

	switch action {
	case admin.AccessAction,
		storage.OpenAction,
		storage.CreateAction,
		storage.PurgeAction,
		presence.ScanAction:
		return true
	}

	return false
}

func (auth *Authorizer) String() string {
	return "mod.user"
}
