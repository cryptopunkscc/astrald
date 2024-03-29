package reflectlink

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/net"
)

type EventLinkReflected struct {
	Link     net.Link
	Endpoint net.Endpoint
}

func (e EventLinkReflected) String() string {
	return fmt.Sprintf("RemoteID=%s ReflectNetwork=%s ReflectAddr=%s",
		e.Link.RemoteIdentity(),
		e.Endpoint.Network(),
		e.Endpoint.String(),
	)
}
