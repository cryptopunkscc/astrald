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
const SetType = "share"
const RemoteSharesSetName = "mod.shares.remote"

type Module interface {
	Authorizer
	AddAuthorizer(Authorizer) error
	RemoveAuthorizer(Authorizer) error

	FindRemoteShare(caller id.Identity, target id.Identity) (RemoteShare, error)
}

type Authorizer interface {
	Authorize(identity id.Identity, dataID data.ID) error
}

type RemoteShare interface {
	Scan(opts *sets.ScanOpts) ([]*sets.Member, error)
	Sync() error
	Unsync() error
	LastUpdate() time.Time
}

var ErrDenied = errors.New("access denied")
