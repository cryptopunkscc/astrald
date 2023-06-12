package tracker

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/events"
	"gorm.io/gorm"
)

const DatabaseName = "tracker.db"
const logTag = "tracker"

// CoreTracker stores information about addresses of other nodes on the network.
type CoreTracker struct {
	db     *gorm.DB
	parser EndpointParser
	events events.Queue
	log    *log.Logger
}

func (tracker *CoreTracker) IdentityByAlias(alias string) (id.Identity, error) {
	var row dbAliases
	if err := tracker.db.First(&row, "alias = ?", alias).Error; err != nil {
		return id.Identity{}, err
	}

	identity, err := id.ParsePublicKeyHex(row.Identity)
	if err != nil {
		return id.Identity{}, err
	}

	return identity, nil
}

// NewCoreTracker returns a new instance of a CoreTracker. It will use db for persistency and the provided unpacker
// to unpack addresses stored in the database.
func NewCoreTracker(assets assets.Store, parser EndpointParser, log *log.Logger, events *events.Queue) (*CoreTracker, error) {
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
	if err = tracker.dbAutoMigrate(); err != nil {
		return nil, err
	}

	if err := tracker.cleanup(); err != nil {
		return nil, err
	}

	return tracker, nil
}

func (tracker *CoreTracker) SetAlias(identity id.Identity, alias string) error {
	return tracker.db.Save(&dbAliases{
		Identity: identity.String(),
		Alias:    alias,
	}).Error
}

func (tracker *CoreTracker) GetAlias(identity id.Identity) (string, error) {
	var row dbAliases
	if err := tracker.db.First(&row, "identity = ?", identity.String()).Error; err != nil {
		return "", err
	}

	return row.Alias, nil
}
