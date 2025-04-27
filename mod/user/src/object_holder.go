package user

import (
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

var _ objects.Holder = &Module{}

func (mod *Module) HoldObject(objectID *object.ID) (hold bool) {
	ac := mod.ActiveContract()
	if ac == nil {
		return false
	}

	return mod.db.AssetsContain(objectID)
}
