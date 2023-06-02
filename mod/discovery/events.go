package discovery

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
)

type EventServicesDiscovered struct {
	Identity id.Identity
	Services []proto.ServiceEntry
}

func (e EventServicesDiscovered) String() string {
	s := log.Sprint("%s", e.Identity)

	for _, srv := range e.Services {
		s = s + fmt.Sprintf(" %s=%s", srv.Name, srv.Type)
	}

	return s
}
