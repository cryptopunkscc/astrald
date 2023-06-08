package discovery

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
)

type EventServicesDiscovered struct {
	identityName string
	Identity     id.Identity
	Services     []proto.ServiceEntry
}

func (e EventServicesDiscovered) String() string {
	s := e.identityName

	for _, srv := range e.Services {
		s = s + fmt.Sprintf(" %s=%s", srv.Name, srv.Type)
	}

	return s
}
