package reflectlink

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/node/link"
)

type EventLinkReflected struct {
	Link *link.Link
	Info *Info
}

func (e EventLinkReflected) String() string {
	return fmt.Sprintf("RemoteID=%s Network=%s LocalEndpoint=%s RemoteEndpoint=%s ReflectAddr=%s",
		e.Link.RemoteIdentity().Fingerprint(),
		e.Link.Network(),
		e.Link.LocalEndpoint(),
		e.Link.RemoteEndpoint(),
		e.Info.ReflectAddr,
	)
}
