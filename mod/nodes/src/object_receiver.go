package nodes

import (
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/mod/utp"
)

func (mod *Module) ReceiveObject(drop objects.Drop) error {
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
	endpoint := event.Endpoint

	var i ip.IP
	switch e := endpoint.(type) {
	case *tcp.Endpoint:
		i = e.IP
	case *utp.Endpoint:
		i = e.IP
	default:
		// unknown endpoint type
		return nil
	}

	if i.IsPublic() {
		mod.log.Log(`nodes module/receiveObservedEndpointMessage observed new public ip: %v`, i)
		mod.AddObservedEndpoint(endpoint, i)
	}

	return nil
}
