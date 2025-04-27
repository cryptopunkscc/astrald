package archives

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

	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity())

	switch object := object.Object.(type) {
	case *objects.EventDiscovered:
		mod.onObjectDiscovered(ctx, object)
	}

	return errors.New("object rejected")
}

func (mod *Module) onObjectDiscovered(ctx *astral.Context, event *objects.EventDiscovered) {
	info, _ := mod.Content.Identify(event.ObjectID)
	if info != nil && info.Type == zipMimeType {
		archive, _ := mod.Index(
			ctx.WithZone(mod.autoIndexZone),
			event.ObjectID,
			nil,
		)

		if archive == nil {
			return
		}

		for _, entry := range archive.Entries {
			mod.Objects.Receive(&objects.EventDiscovered{
				ObjectID: entry.ObjectID,
				Zone:     astral.ZoneVirtual | event.Zone,
			}, nil)
		}
	}
	return
}
