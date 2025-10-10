package nodes

import (
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/mod/utp"
)

func (mod *Module) ReceiveObject(drop objects.Drop) error {
	// only receive objects from the local node

	// FIXME: rate trustability of sender

	switch object := drop.Object().(type) {
	case *nodes.ObservedEndpointMessage:
		err := mod.receiveObservedEndpointMessage(object)
		if err == nil {
			return drop.Accept(false)
		}
	}

	return nil
}

func (mod *Module) receiveObservedEndpointMessage(event *nodes.ObservedEndpointMessage) error {
	// FIXME: 3 last unique map entries with timestamps
	var i ip.IP
	switch e := event.Endpoint.(type) {
	case *tcp.Endpoint:
		i = e.IP
	case *utp.Endpoint:
		i = e.IP
	default:
		// unknown endpoint type
		return nil
	}

	if i.IsGlobalUnicast() {
		mod.log.Log(`observed new public IP: %v`, i)
		mod.lastObservedIP = i
	}

	return nil
}
