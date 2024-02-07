package shares

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/sets"
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
	LocalShare(caller id.Identity, create bool) (LocalShare, error)
	Notify(identity id.Identity) error
}

type Authorizer interface {
	Authorize(identity id.Identity, dataID data.ID) error
}

type LocalShare interface {
	Identity() id.Identity
	AddSet(...string) error
	RemoveSet(...string) error
	AddObject(...data.ID) error
	RemoveObject(...data.ID) error
	Scan(opts *sets.ScanOpts) ([]*sets.Member, error)
}

type RemoteShare interface {
	Sync() error
	Unsync() error
	LastUpdate() time.Time
}

var ErrDenied = errors.New("access denied")
