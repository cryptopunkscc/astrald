package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

type EventCommitted struct {
	ObjectID object.ID
}

type EventPurged struct {
	ObjectID object.ID
}

type EventDiscovered struct {
	ObjectID object.ID
	Zone     astral.Zone
}

type EventHeld struct {
	HolderID  *astral.Identity
	ObjectIDs []object.ID
}

type EventReleased struct {
	HolderID  *astral.Identity
	ObjectIDs []object.ID
}
