package objects

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/object"
)

type EventObjectCommitted struct {
	ObjectID object.ID
}

type EventObjectPurged struct {
	ObjectID object.ID
}

type EventObjectDiscovered struct {
	ObjectID object.ID
	Zone     net.Zone
}
