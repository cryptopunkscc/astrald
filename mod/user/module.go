package user

import (
	"github.com/cryptopunkscc/astrald/auth/id"
)

const ModuleName = "user"

type Module interface {
	AddIdentity(identity id.Identity) error
	RemoveIdentity(identity id.Identity) error
	Identities() []id.Identity
}
