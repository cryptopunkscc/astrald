package notify

import (
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
)

type Notify func([]Notification)

type Notification struct {
	warpdrive.Peer
	warpdrive.Offer
	*warpdrive.Info
}
