package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
)

const ModuleName = "auth"

type Module interface {
	Authorize(identity id.Identity, action string, target astral.Object) bool
	AddAuthorizer(Authorizer) error
}

type Authorizer interface {
	Authorize(identity id.Identity, action string, target astral.Object) bool
}
