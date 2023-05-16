package tracker

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/event"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"path/filepath"
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
func NewCoreTracker(configStore config.Store, parser EndpointParser, log *log.Logger, events *event.Queue) (*CoreTracker, error) {
	tracker := &CoreTracker{
		parser: parser,
		log:    log.Tag(logTag),
	}

	tracker.events.SetParent(events)

	var err error
	var path string

	if fileStore, ok := configStore.(*config.FileStore); ok {
		path = fileStore.BaseDir()
	}

	tracker.db, err = gorm.Open(
		sqlite.Open(filepath.Join(path, DatabaseName)),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
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
