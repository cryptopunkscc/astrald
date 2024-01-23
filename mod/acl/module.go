package acl

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
)

const ModuleName = "acl"

type Module interface {
	Grant(identity id.Identity, dataID data.ID) error
	Revoke(identity id.Identity, dataID data.ID) error
}
