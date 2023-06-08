package discovery

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
)

type EventServicesDiscovered struct {
	identityName string
	Identity     id.Identity
	Services     []ServiceEntry
}

func (e EventServicesDiscovered) String() string {
	s := e.identityName

	for _, srv := range e.Services {
		s = s + fmt.Sprintf(" %s=%s", srv.Name, srv.Type)
	}

	return s
}
