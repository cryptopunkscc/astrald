package user

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/authorizer"
)

var _ authorizer.Authorizer = &Authorizer{}

type Authorizer struct {
	mod *Module
}

func (auth *Authorizer) Authorize(identity id.Identity, action string, args ...any) bool {
	_, found := auth.mod.identities.Get(identity.PublicKeyHex())
	if !found {
		return false
	}

	switch action {
	case admin.AccessAction,
		storage.OpenAction,
		storage.CreateAction:
		return true
	}

	return false
}

func (auth *Authorizer) String() string {
	return "mod.user"
}
