package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/user"
)

var _ objects.Holder = &Module{}

func (mod *Module) HoldObject(objectID *astral.ObjectID) (hold bool) {
	ac := mod.ActiveContract()
	if ac == nil {
		return false
	}

	if mod.db.AssetsContain(objectID) {
		return true
	}

	objectType, err := mod.Objects.GetType(mod.ctx, objectID)
	if err != nil {
		mod.log.Error("failed to get object type of %v", objectID)
		return false
	}

	switch objectType {
	case user.SignedNodeContract{}.ObjectType():
		return mod.holdNodeContract(objectID)
	case user.SignedNodeContractRevocation{}.ObjectType():
		return mod.holdNodeContractRevocation(objectID)
	default:
		return false
	}
}

func (mod *Module) holdNodeContract(objectID *astral.ObjectID) bool {
	c, err := mod.FindNodeContract(objectID)
	if err != nil {
		mod.log.Error("failed to load node contract %v: %v", objectID, err)
		return false
	}

	if c.IsExpired() {
		return false
	}

	return true
}

func (mod *Module) holdNodeContractRevocation(objectID *astral.ObjectID) bool {
	c, err := mod.FindNodeContractRevocation(objectID)
	if err != nil {
		mod.log.Error("failed to load node contract revocation %v: %v", objectID, err)
		return false
	}

	if c.IsExpired() {
		return false
	}

	return true
}
