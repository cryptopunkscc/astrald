package reflectlink

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
)

type EventLinkReflected struct {
	Link     *link.Link
	Endpoint net.Endpoint
}

func (e EventLinkReflected) String() string {
	return fmt.Sprintf("RemoteID=%s ReflectNetwork=%s ReflectAddr=%s",
		e.Link.RemoteIdentity(),
		e.Endpoint.Network(),
		e.Endpoint.String(),
	)
}
