package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
)

type DataAuthorizer interface {
	Authorize(identity id.Identity, dataID data.ID) error
}
