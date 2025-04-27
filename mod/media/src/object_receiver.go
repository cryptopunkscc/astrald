package media

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

var _ objects.Receiver = &Module{}

func (mod *Module) ReceiveObject(object *objects.SourcedObject) error {
	// only receive objects from the local node
	if !object.Source.IsEqual(mod.node.Identity()) {
		return errors.New("object rejected")
	}

	switch object := object.Object.(type) {
	case *objects.EventDiscovered:
		mod.receiveObjectDiscovered(object)
	}

	return errors.New("object rejected")
}

func (mod *Module) receiveObjectDiscovered(event *objects.EventDiscovered) {
	mod.DescribeObject(astral.NewContext(nil).WithIdentity(mod.node.Identity()), event.ObjectID, astral.DefaultScope())
}
