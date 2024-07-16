package reflectlink

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/net"
)

type EventLinkReflected struct {
	Link     net.Link
	Endpoint exonet.Endpoint
}

func (e EventLinkReflected) String() string {
	return fmt.Sprintf("RemoteID=%s ReflectNetwork=%s ReflectAddr=%s",
		e.Link.RemoteIdentity(),
		e.Endpoint.Network(),
		e.Endpoint.Address(),
	)
}
