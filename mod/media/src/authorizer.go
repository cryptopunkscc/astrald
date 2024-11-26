package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) Authorize(identity *astral.Identity, action string, target astral.Object) bool {
	switch action {
	case objects.ActionReadDescriptor:
		switch target.ObjectType() {
		case media.AudioDescriptor{}.ObjectType():
			return true
		}

	case objects.ActionRead:
		if target == nil {
			return false
		}

		objectID, ok := target.(*object.ID)
		if !ok {
			return false
		}

		parentID := mod.getParentID(*objectID)
		if parentID.IsZero() {
			return false
		}

		return mod.Auth.Authorize(identity, action, &parentID)
	}

	return false
}
