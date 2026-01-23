package user

import (
	"errors"

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

	return mod.holdNodeContract(objectID) || mod.holdNodeContractRevocation(objectID)
}

func (mod *Module) holdNodeContract(objectID *astral.ObjectID) bool {
	c, err := mod.FindNodeContract(objectID)
	if err != nil {
		if errors.Is(err, user.ErrNodeContractNotFound) {
			return false
		}

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
		if errors.Is(err, user.ErrContractRevocationNotFound) {
			return false
		}
		mod.log.Error("failed to load node contract revocation %v: %v", objectID, err)
		return false
	}

	if c.IsExpired() {
		return false
	}

	return true
}
