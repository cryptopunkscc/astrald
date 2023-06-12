package tracker

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

func (tracker *CoreTracker) EndpointsByIdentity(identity id.Identity) ([]net.Endpoint, error) {
	var dbEndpoints []dbEndpoint

	if tx := tracker.db.Find(
		&dbEndpoints,
		"identity = ? and expires_at > ?",
		identity.String(),
		time.Now(),
	); tx.Error != nil {
		return nil, tx.Error
	}

	var endpoints = make([]net.Endpoint, 0, len(dbEndpoints))

	for _, dbEp := range dbEndpoints {
		if ep, err := tracker.dbEndpointToEndopoint(dbEp); err == nil {
			endpoints = append(endpoints, ep)
		}
	}

	return endpoints, nil
}

func (tracker *CoreTracker) find(identity id.Identity, e net.Endpoint) (dbEp dbEndpoint, err error) {
	err = tracker.db.First(&dbEp, dbEndpoint{
		Identity: identity.String(),
		Network:  e.Network(),
		Address:  e.String(),
	}).Error
	return
}
