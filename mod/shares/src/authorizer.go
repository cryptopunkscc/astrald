package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/object"
)

var _ node.Authorizer = &Authorizer{}

type Authorizer struct {
	mod *Module
}

func (auth *Authorizer) Authorize(identity id.Identity, action string, args ...any) bool {
	switch action {
	case objects.ActionRead:
		if len(args) == 0 {
			return false
		}
		objectID, ok := args[0].(object.ID)
		if !ok {
			return false
		}

		return auth.mod.Authorize(identity, objectID) == nil
	}

	return false
}

func (auth *Authorizer) String() string {
	return "mod.shares"
}
