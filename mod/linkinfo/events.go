package linkinfo

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/node/link"
)

type EventLinkInfo struct {
	Link *link.Link
	Info *Info
}

func (e EventLinkInfo) String() string {
	return fmt.Sprintf("received link info from %s", e.Link.RemoteIdentity().Fingerprint())
}
