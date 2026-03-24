package archives

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) Authorize(identity *astral.Identity, action auth.Action, target astral.Object) bool {
	return auth.Auth(auth.ActionsMap{
		objects.ActionRead: {auth.NewHandler(mod.AuthorizeRead)},
	}, identity, action, target)
}

func (mod *Module) AuthorizeRead(identity *astral.Identity, objectID *astral.ObjectID) bool {
	var rows []*dbEntry

	var err = mod.db.
		Unscoped().
		Preload("Parent").
		Where("object_id = ?", objectID).
		Find(&rows).Error
	if err != nil {
		return false
	}

	for _, row := range rows {
		if row.Parent == nil {
			mod.log.Errorv(1, "db: entry for %v references an invalid parent", objectID)
			continue
		}

		zipID := row.Parent.ObjectID

		// sanity check
		if zipID.IsEqual(objectID) {
			continue
		}

		return mod.Auth.Authorize(identity, objects.ActionRead, zipID)
	}

	return false
}
