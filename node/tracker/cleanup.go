package tracker

import (
	"time"
)

// cleanup cleans the database of expired addresses
func (tracker *CoreTracker) cleanup() error {
	return tracker.db.Delete(&dbEndpoint{}, "expires_at < ?", time.Now()).Error
}
