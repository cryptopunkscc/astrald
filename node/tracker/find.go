package tracker

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

func (tracker *CoreTracker) EndpointsByIdentity(identity id.Identity) ([]net.Endpoint, error) {
	var rows []dbEndpoint

	if err := tracker.db.Find(&rows, "identity = ?", identity.String()).Error; err != nil {
		return nil, err
	}

	var endpoints = make([]net.Endpoint, 0, len(rows))

	for _, dbEp := range rows {
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
