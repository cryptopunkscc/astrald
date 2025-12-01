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

	objectType, err := mod.Objects.GetType(mod.ctx, objectID)
	if err != nil {
		mod.log.Error("failed to get object type of %v", objectID)
		return false
	}

	switch objectType {
	case user.SignedNodeContract{}.ObjectType():
		c, err := mod.FindNodeContract(objectID)
		if err != nil {
			mod.log.Error("failed to load node contract %v: %v", objectID, err)
			return false
		}

		if c.IsExpired() {
			return false
		}

		return true
	case user.SignedNodeContractRevocation{}.ObjectType():
		c, err := mod.FindNodeContractRevocation(objectID)
		if err != nil {
			mod.log.Error("failed to load node contract revocation %v: %v", objectID, err)
			return false
		}

		if c.IsExpired() {
			return false
		}

		return true
	default:
		return mod.db.AssetsContain(objectID)
	}
}
