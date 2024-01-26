package tracker

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/events"
	"gorm.io/gorm"
)

const DatabaseName = "tracker.db"
const logTag = "tracker"

var _ Tracker = &CoreTracker{}

// CoreTracker stores information about addresses of other nodes on the network.
type CoreTracker struct {
	db     *gorm.DB
	parser EndpointParser
	events events.Queue
	log    *log.Logger
}

type EndpointParser interface {
	Parse(network string, address string) (net.Endpoint, error)
}

// NewCoreTracker returns a new instance of a CoreTracker. It will use db for persistency and the provided unpacker
// to unpack addresses stored in the database.
func NewCoreTracker(assets assets.Assets, parser EndpointParser, log *log.Logger, events *events.Queue) (*CoreTracker, error) {
	var err error
	var tracker = &CoreTracker{
		parser: parser,
		log:    log.Tag(logTag),
	}

	tracker.events.SetParent(events)

	tracker.db = assets.Database()

	// Migrate the schema
	if err = tracker.dbAutoMigrate(); err != nil {
		return nil, err
	}

	return tracker, nil
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

// SetAlias sets the alias for the identity. Set an empty alias to unset.
func (tracker *CoreTracker) SetAlias(identity id.Identity, alias string) error {
	if alias == "" {
		return tracker.db.Delete(&dbAliases{}, "identity = ?", identity.String()).Error
	}

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

// Clear deletes all enpoints of the identity
func (tracker *CoreTracker) Clear(identity id.Identity) error {
	return tracker.db.Delete(&dbEndpoint{}, "identity = ?", identity.String()).Error
}

// Remove removes enpoints and the alias of the identity
func (tracker *CoreTracker) Remove(identity id.Identity) error {
	if err := tracker.db.Delete(&dbAliases{}, "identity = ?", identity.String()).Error; err != nil {
		return err
	}

	return tracker.db.Delete(&dbEndpoint{}, "identity = ?", identity.String()).Error
}
