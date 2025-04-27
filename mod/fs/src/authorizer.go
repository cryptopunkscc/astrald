package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"strings"
)

func (mod *Module) Authorize(identity *astral.Identity, action auth.Action, target astral.Object) bool {
	switch action {
	case objects.ActionRead:
		if id, ok := target.(*object.ID); ok {
			return mod.authorizeRead(identity, id)
		}
		return false

	default:
		return false
	}
}

func (mod *Module) authorizeRead(identity *astral.Identity, objectID *object.ID) bool {
	for _, path := range mod.Path(objectID) {
		for _, p := range mod.shares.Keys() {
			if strings.HasPrefix(path, p) {
				a, _ := mod.shares.Get(p)
				if a.Contains(identity.String()) || a.Contains("anyone") {
					return true
				}
			}
		}
	}

	return false
}
