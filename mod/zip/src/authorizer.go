package zip

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/authorizer"
)

var _ authorizer.Authorizer = &Authorizer{}

type Authorizer struct {
	mod *Module
}

func (auth *Authorizer) Authorize(identity id.Identity, action string, args ...any) bool {
	switch action {
	case storage.OpenAction:
		if len(args) == 0 {
			return false
		}
		dataID, ok := args[0].(data.ID)
		if !ok {
			return false
		}

		var rows []*dbContents

		var err = auth.mod.db.
			Unscoped().
			Preload("Zip").
			Where("file_id = ?", dataID).
			Find(&rows).Error
		if err != nil {
			return false
		}

		for _, row := range rows {
			if row.Zip == nil {
				auth.mod.log.Errorv(1, "db row for file %v has null reference to zip", dataID)
				continue
			}

			zipID := row.Zip.DataID

			// sanity check
			if zipID.IsEqual(dataID) {
				continue
			}

			return auth.mod.node.Auth().Authorize(identity, storage.OpenAction, zipID)
		}
	}

	return false
}

func (auth *Authorizer) String() string {
	return "mod.zip"
}
