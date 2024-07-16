package objects

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/net"
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
	Zone     net.Zone
}

type EventHeld struct {
	HolderID  id.Identity
	ObjectIDs []object.ID
}

type EventReleased struct {
	HolderID  id.Identity
	ObjectIDs []object.ID
}
