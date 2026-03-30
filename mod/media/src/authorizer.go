package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) AuthorizeObjectsReadByFile(ctx *astral.Context, identity *astral.Identity, f *media.AudioFile) bool {
	return mod.Auth.Authorize(ctx, identity, objects.ActionRead, f.ObjectID)
}

func (mod *Module) AuthorizeObjectsReadByID(ctx *astral.Context, identity *astral.Identity, objectID *astral.ObjectID) bool {
	parentID, err := mod.db.FindAudioContainerID(objectID)
	if err != nil || parentID.IsZero() {
		return false
	}
	return mod.Auth.Authorize(ctx, identity, objects.ActionRead, parentID)
}
