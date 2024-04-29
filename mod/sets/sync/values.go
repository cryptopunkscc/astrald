package sync

import (
	"errors"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

const (
	opDone   = 0x00
	opAdd    = 0x01
	opRemove = 0x02
	opResync = 0x03
)

type Diff struct {
	Updates []Update
	Time    time.Time
}

type Update struct {
	ObjectID object.ID
	Present  bool
}

var ErrResyncRequired = errors.New("resync required")
var ErrProtocolError = errors.New("protocol error")
