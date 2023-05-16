package tracker

import "github.com/cryptopunkscc/astrald/auth/id"

func (tracker *CoreTracker) DeleteAll(identity id.Identity) error {
	return tracker.db.Delete(&dbEndpoint{}, "identity = ?", identity.String()).Error
}
