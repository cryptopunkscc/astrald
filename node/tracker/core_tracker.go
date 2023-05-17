package tracker

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/event"
	"gorm.io/gorm"
)

const DatabaseName = "tracker.db"
const logTag = "tracker"

// CoreTracker stores information about addresses of other nodes on the network.
type CoreTracker struct {
	db     *gorm.DB
	parser EndpointParser
	events event.Queue
	log    *log.Logger
}

// NewCoreTracker returns a new instance of a CoreTracker. It will use db for persistency and the provided unpacker
// to unpack addresses stored in the database.
func NewCoreTracker(assets assets.Store, parser EndpointParser, log *log.Logger, events *event.Queue) (*CoreTracker, error) {
	var err error
	var tracker = &CoreTracker{
		parser: parser,
		log:    log.Tag(logTag),
	}

	tracker.events.SetParent(events)

	tracker.db, err = assets.OpenDB(DatabaseName)
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	if err := tracker.db.AutoMigrate(&dbEndpoint{}); err != nil {
		return nil, err
	}

	if err := tracker.cleanup(); err != nil {
		return nil, err
	}

	return tracker, nil
}
