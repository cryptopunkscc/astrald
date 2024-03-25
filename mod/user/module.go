package user

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
)

const ModuleName = "user"
const DBPrefix = "users__"

type Module interface {
	LocalUser() LocalUser
	SetLocalUser(identity id.Identity) error
}

type Profile struct {
	Identity id.Identity
	Name     string
	Avatar   data.ID
}

type LocalUser interface {
	Identity() id.Identity
}

type UserDesc struct {
	Name string
}

func (UserDesc) Type() string {
	return "user"
}
