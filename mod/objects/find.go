package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

// Finder is used to figure out which identities can provide access to an object
type Finder interface {
	FindObject(*astral.Context, *object.ID, *astral.Scope) []*astral.Identity
}
