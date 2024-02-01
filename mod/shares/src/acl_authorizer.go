package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
)

type ACLAuthorizer struct {
	*Module
}

func (auth *ACLAuthorizer) Authorize(identity id.Identity, dataID data.ID) error {
	share, err := auth.FindShare(identity)
	if err != nil {
		return shares.ErrDenied
	}

	scan, err := share.Scan(&sets.ScanOpts{DataID: dataID})
	if len(scan) == 1 {
		return nil
	}

	return shares.ErrDenied
}
