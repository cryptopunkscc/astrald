package tracker

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

// AddEndpoint adds an endpoint to the identity. If the endpoint already exists, its expiry time will be replaced.
func (tracker *CoreTracker) AddEndpoint(identity id.Identity, e net.Endpoint) (err error) {
	var dbEp dbEndpoint

	e, err = tracker.parser.Parse(e.Network(), e.String())
	if err != nil {
		return
	}

	if dbEp, err = tracker.find(identity, e); err != nil {
		err = tracker.db.Create(&dbEndpoint{
			Identity: identity.String(),
			Network:  e.Network(),
			Address:  e.String(),
		}).Error

		if err == nil {
			tracker.events.Emit(EventNewEndpoint{
				Identity: identity,
				Endpoint: e,
			})
		}
		return
	}

	return tracker.db.Save(&dbEp).Error
}
