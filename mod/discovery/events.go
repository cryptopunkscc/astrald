package discovery

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/id"
)

type EventDiscovered struct {
	IdentitiyName string
	Identity      id.Identity
	Info          *Info
}

func NewEventDiscovered(identityName string, identity id.Identity, info *Info) EventDiscovered {
	return EventDiscovered{
		IdentitiyName: identityName,
		Identity:      identity,
		Info:          info,
	}
}

func (e EventDiscovered) String() string {
	s := e.IdentitiyName

	if s == "" {
		s = e.Identity.Fingerprint()
	}

	for _, srv := range e.Info.Services {
		s = s + fmt.Sprintf(" %s=%s", srv.Name, srv.Type)
	}

	return s
}
