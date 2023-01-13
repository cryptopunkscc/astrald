package tracker

import (
	"github.com/cryptopunkscc/astrald/node/db"
	"github.com/cryptopunkscc/astrald/node/event"
)

// Tracker stores information about addresses of other nodes on the network.
type Tracker struct {
	db       *db.Database
	unpacker AddrUnpacker
	events   event.Queue
}

// New returns a new instance of a Tracker. It will use db for persistency and the provided unpacker to unpack
// addresses stored in the database.
func New(db *db.Database, unpacker AddrUnpacker) (*Tracker, error) {
	tracker := &Tracker{db: db, unpacker: unpacker}

	if err := tracker.purge(); err != nil {
		return nil, err
	}

	return tracker, nil
}
