package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

var _ objects.Holder = &Module{}

func (mod *Module) HoldObject(objectID *astral.ObjectID) (hold bool) {
	switch {
	case mod.db.assetExists(objectID):
		// hold all user assets
		return true

	case mod.db.signedNodeContractExists(objectID):
		// hold all indexed signed node contracts
		return true

	case mod.db.nodeContractRevocationExists(objectID):
		// hold all indexed contract revocations
		return true
	}

	return false
}
