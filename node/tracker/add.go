package tracker

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

// Add adds an endpoint to the identity. If the endpoint already exists, its expiry time will be replaced.
func (tracker *CoreTracker) Add(identity id.Identity, e net.Endpoint, expiresAt time.Time) (err error) {
	var dbEp dbEndpoint

	e, err = tracker.parser.Parse(e.Network(), e.String())
	if err != nil {
		return
	}

	if dbEp, err = tracker.find(identity, e); err != nil {
		err = tracker.db.Create(&dbEndpoint{
			Identity:  identity.String(),
			Network:   e.Network(),
			Address:   e.String(),
			ExpiresAt: expiresAt,
		}).Error

		if err == nil {
			tracker.events.Emit(EventNewEndpoint{
				TrackedEndpoint: TrackedEndpoint{
					Identity:  identity,
					Endpoint:  e,
					ExpiresAt: expiresAt,
				},
			})
		}
		return
	}

	if expiresAt.After(dbEp.ExpiresAt) {
		dbEp.ExpiresAt = expiresAt
	}

	return tracker.db.Save(&dbEp).Error
}
