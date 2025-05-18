package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

var _ objects.Holder = &Module{}

func (mod *Module) HoldObject(objectID *astral.ObjectID) (hold bool) {
	ac := mod.ActiveContract()
	if ac == nil {
		return false
	}

	return mod.db.AssetsContain(objectID)
}
