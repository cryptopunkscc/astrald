package nodes

import (
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) ReceiveObject(drop objects.Drop) error {
	// only receive objects from the local node

	// FIXME: rate trustability of sender
	switch object := drop.Object().(type) {
	case *nodes.ObservedEndpointEvent:
		err := mod.receiveObservedEndpointEvent(object)
		if err == nil {
			return drop.Accept(false)
		}
	}

	return nil
}

func (mod *Module) receiveObservedEndpointEvent(event *nodes.ObservedEndpointEvent) error {
	// FIXME: implement

	return nil
}
