package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/authorizer"
)

var _ authorizer.Authorizer = &Authorizer{}

type Authorizer struct {
	mod *Module
}

func (auth *Authorizer) Authorize(identity id.Identity, action string, args ...any) bool {
	switch action {
	case storage.OpenAction:
		if len(args) == 0 {
			return false
		}
		dataID, ok := args[0].(data.ID)
		if !ok {
			return false
		}

		return auth.mod.Authorize(identity, dataID) == nil
	}

	return false
}

func (auth *Authorizer) String() string {
	return "mod.shares"
}
