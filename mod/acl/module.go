package acl

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

const ModuleName = "acl"

type Module interface {
	Grant(identity id.Identity, dataID data.ID, expiresAt time.Time) error
	Revoke(identity id.Identity, dataID data.ID) error
	Verify(identity id.Identity, dataID data.ID) bool
}
