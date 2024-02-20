package user

import (
	"github.com/cryptopunkscc/astrald/auth/id"
)

const ModuleName = "user"
const DBPrefix = "user__"

type Module interface {
	AddIdentity(identity id.Identity) error
	RemoveIdentity(identity id.Identity) error
	Identities() []id.Identity
}

type UserDesc struct {
	Name string
}

func (UserDesc) Type() string {
	return "user"
}
