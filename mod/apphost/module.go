package apphost

import (
	"github.com/cryptopunkscc/astrald/auth/id"
)

const ModuleName = "apphost"

type Module interface {
	SetDefaultIdentity(id.Identity) error
	DefaultIdentity() id.Identity
	CreateAccessToken(identity id.Identity) (string, error)
}
