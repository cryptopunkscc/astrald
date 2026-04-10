package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) AuthorizeObjectsRead(ctx *astral.Context, action *objects.ReadObjectAction) bool {
	parentID, err := mod.db.FindAudioContainerID(action.ObjectID)
	if err != nil || parentID.IsZero() {
		return false
	}

	return mod.Auth.Authorize(ctx, &objects.ReadObjectAction{
		Action:   auth.NewAction(action.Actor()),
		ObjectID: parentID,
	})
}
