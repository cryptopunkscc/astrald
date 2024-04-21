package sync

import (
	"errors"
	"github.com/cryptopunkscc/astrald/data"
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
	DataID  data.ID
	Present bool
}

var ErrResyncRequired = errors.New("resync required")
var ErrProtocolError = errors.New("protocol error")
