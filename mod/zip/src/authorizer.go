package zip

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/node/authorizer"
	"github.com/cryptopunkscc/astrald/object"
)

var _ authorizer.Authorizer = &Authorizer{}

type Authorizer struct {
	mod *Module
}

func (auth *Authorizer) Authorize(identity id.Identity, action string, args ...any) bool {
	switch action {
	case objects.ReadAction:
		if len(args) == 0 {
			return false
		}
		objectID, ok := args[0].(object.ID)
		if !ok {
			return false
		}

		var rows []*dbContents

		var err = auth.mod.db.
			Unscoped().
			Preload("Zip").
			Where("file_id = ?", objectID).
			Find(&rows).Error
		if err != nil {
			return false
		}

		for _, row := range rows {
			if row.Zip == nil {
				auth.mod.log.Errorv(1, "db row for file %v has null reference to zip", objectID)
				continue
			}

			zipID := row.Zip.DataID

			// sanity check
			if zipID.IsEqual(objectID) {
				continue
			}

			return auth.mod.node.Auth().Authorize(identity, objects.ReadAction, zipID)
		}
	}

	return false
}

func (auth *Authorizer) String() string {
	return "mod.zip"
}
