package sdp

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
)

type EventServicesDiscovered struct {
	identityName string
	Identity     id.Identity
	Services     []ServiceEntry
}

func NewEventServicesDiscovered(identityName string, identity id.Identity, services []ServiceEntry) EventServicesDiscovered {
	return EventServicesDiscovered{
		identityName: identityName,
		Identity:     identity,
		Services:     services,
	}
}

func (e EventServicesDiscovered) String() string {
	s := e.identityName

	if s == "" {
		s = e.Identity.Fingerprint()
	}

	for _, srv := range e.Services {
		s = s + fmt.Sprintf(" %s=%s", srv.Name, srv.Type)
	}

	return s
}
