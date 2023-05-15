package tracker

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/db"
	"github.com/cryptopunkscc/astrald/node/event"
)

// CoreTracker stores information about addresses of other nodes on the network.
type CoreTracker struct {
	db       *db.Database
	unpacker AddrUnpacker
	events   event.Queue
	log      *log.Logger
}

// NewCoreTracker returns a new instance of a CoreTracker. It will use db for persistency and the provided unpacker
// to unpack addresses stored in the database.
func NewCoreTracker(db *db.Database, unpacker AddrUnpacker, log *log.Logger) (*CoreTracker, error) {
	tracker := &CoreTracker{
		db:       db,
		unpacker: unpacker,
		log:      log.Tag(logTag),
	}

	if err := tracker.purge(); err != nil {
		return nil, err
	}

	return tracker, nil
}
