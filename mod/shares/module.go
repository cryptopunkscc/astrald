package shares

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

const ModuleName = "shares"
const DBPrefix = "shares__"
const RemoteSetType = "remote"
const LocalSetType = "local"

type Module interface {
	Authorizer
	AddAuthorizer(Authorizer) error
	RemoveAuthorizer(Authorizer) error

	FindRemoteShare(caller id.Identity, target id.Identity) (RemoteShare, error)
	Notify(identity id.Identity) error
}

type Authorizer interface {
	Authorize(identity id.Identity, dataID data.ID) error
}

type RemoteShare interface {
	Sync() error
	Unsync() error
	LastUpdate() time.Time
}

var ErrDenied = errors.New("access denied")
