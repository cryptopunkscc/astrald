package nat

import (
	"github.com/cryptopunkscc/astrald/mod/events"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) ReceiveObject(drop objects.Drop) error {
	switch object := drop.Object().(type) {
	case *events.Event:
		switch object.Data.(type) {
		case *nodes.ObservedEndpointChangedEvent:
			mod.evaluateEnabled()
		}
	}

	return nil
}
