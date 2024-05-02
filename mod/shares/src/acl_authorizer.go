package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/object"
)

type ACLAuthorizer struct {
	*Module
}

func (auth *ACLAuthorizer) Authorize(identity id.Identity, objectID object.ID) error {
	set, err := auth.openExportSet(identity)
	if err != nil {
		return shares.ErrDenied
	}

	scan, err := set.Scan(&sets.ScanOpts{ObjectID: objectID})
	if len(scan) == 1 {
		return nil
	}

	return shares.ErrDenied
}
