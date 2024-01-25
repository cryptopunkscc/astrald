package shares

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

const ModuleName = "shares"

type Module interface {
	Authorizer
	AddAuthorizer(Authorizer) error
	RemoveAuthorizer(Authorizer) error

	Grant(identity id.Identity, dataID data.ID) error
	Revoke(identity id.Identity, dataID data.ID) error

	Sync(caller id.Identity, target id.Identity) error
	LastSynced(caller id.Identity, target id.Identity) (time.Time, error)
}

type Authorizer interface {
	Authorize(identity id.Identity, dataID data.ID) error
}

type RemoteShare struct {
	Caller id.Identity
	Target id.Identity
}

var ErrDenied = errors.New("access denied")
