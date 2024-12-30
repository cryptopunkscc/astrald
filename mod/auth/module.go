package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "auth"

type Module interface {
	Authorize(identity *astral.Identity, action Action, target astral.Object) bool
	AddAuthorizer(Authorizer) error
}

type Authorizer interface {
	Authorize(identity *astral.Identity, action Action, target astral.Object) bool
}
