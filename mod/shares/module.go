package shares

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
)

const ModuleName = "shares"

type Module interface {
	Authorizer
	AddAuthorizer(Authorizer) error
	RemoveAuthorizer(Authorizer) error

	Grant(identity id.Identity, dataID data.ID) error
	Revoke(identity id.Identity, dataID data.ID) error
}

type Authorizer interface {
	Authorize(identity id.Identity, dataID data.ID) error
}

var ErrDenied = errors.New("access denied")
