package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) Authorize(identity *astral.Identity, action auth.Action, target astral.Object) bool {
	return auth.Auth(auth.ActionsMap{
		objects.ActionRead: {auth.NewHandler(mod.AuthorizeAudioFile), auth.NewHandler(mod.AuthorizeObjectID)},
	}, identity, action, target)
}

func (mod *Module) AuthorizeAudioFile(identity *astral.Identity, f *media.AudioFile) bool {
	return mod.Auth.Authorize(identity, objects.ActionRead, f.ObjectID)
}

func (mod *Module) AuthorizeObjectID(identity *astral.Identity, objectID *astral.ObjectID) bool {
	parentID, err := mod.db.FindAudioContainerID(objectID)
	if err != nil || parentID.IsZero() {
		return false
	}
	return mod.Auth.Authorize(identity, objects.ActionRead, parentID)
}
