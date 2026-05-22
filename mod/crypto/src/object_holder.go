package crypto

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

var _ objects.Holder = &Module{}

func (mod *Module) HoldObject(objectID *astral.ObjectID) bool {
	held, err := mod.db.isKeyIndexed(objectID)
	if err != nil {
		mod.log.Error("object hold lookup failed: %v", err)
		return true
	}
	return held
}
