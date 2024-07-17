package archives

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

var _ auth.Authorizer = &Authorizer{}

type Authorizer struct {
	mod *Module
}

func (auth *Authorizer) Authorize(identity id.Identity, action string, args ...any) bool {
	switch action {
	case objects.ActionRead:
		if len(args) == 0 {
			return false
		}
		objectID, ok := args[0].(object.ID)
		if !ok {
			return false
		}

		var rows []*dbEntry

		var err = auth.mod.db.
			Unscoped().
			Preload("Parent").
			Where("object_id = ?", objectID).
			Find(&rows).Error
		if err != nil {
			return false
		}

		for _, row := range rows {
			if row.Parent == nil {
				auth.mod.log.Errorv(1, "db: entry for %v references an invalid parent", objectID)
				continue
			}

			zipID := row.Parent.ObjectID

			// sanity check
			if zipID.IsEqual(objectID) {
				continue
			}

			return auth.mod.auth.Authorize(identity, objects.ActionRead, zipID)
		}
	}

	return false
}

func (auth *Authorizer) String() string {
	return "mod.archives"
}
