package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) Authorize(identity *astral.Identity, action auth.Action, target astral.Object) bool {
	switch action {
	case objects.ActionRead:
		if target == nil {
			return false
		}

		switch target := target.(type) {
		case *media.AudioFile:
			return mod.Auth.Authorize(identity, action, target.ObjectID)

		case *object.ID:
			parentID, err := mod.db.FindAudioContainerID(target)
			if err != nil || parentID.IsZero() {
				return false
			}

			return mod.Auth.Authorize(identity, action, parentID)
		}
	}

	return false
}
