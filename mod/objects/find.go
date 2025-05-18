package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
)

// Finder is used to figure out which identities can provide access to an object
type Finder interface {
	FindObject(*astral.Context, *astral.ObjectID, *astral.Scope) []*astral.Identity
}
