package apphost

import "github.com/cryptopunkscc/astrald/astral"

func (mod *Module) HoldObject(objectID *astral.ObjectID) bool {
	held, err := mod.db.ObjectHeld(objectID)
	if err != nil {
		mod.log.Error("object hold lookup failed: %v", err)
		return true
	}
	return held
}
