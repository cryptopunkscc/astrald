package shares

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/object"
)

type DataAuthorizer interface {
	Authorize(id.Identity, object.ID) error
}
