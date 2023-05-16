package tracker

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"time"
)

// Identities returns a list of all tracked identities
func (tracker *CoreTracker) Identities() ([]id.Identity, error) {
	type row struct {
		Identity string
	}

	var dbIDs []row

	err := tracker.db.
		Model(&dbEndpoint{}).
		Select("identity").
		Group("identity").
		Find(&dbIDs, "expires_at > ?", time.Now()).Error
	if err != nil {
		return nil, err
	}

	identities := make([]id.Identity, 0)

	for _, dbID := range dbIDs {
		identity, err := id.ParsePublicKeyHex(dbID.Identity)
		if err != nil {
			continue
		}
		identities = append(identities, identity)
	}

	return identities, nil
}
