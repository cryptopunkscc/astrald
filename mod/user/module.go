package user

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/object"
)

const ModuleName = "user"
const DBPrefix = "users__"

type Module interface {
	UserID() id.Identity
	SetUserID(userID id.Identity) error
}

type Profile struct {
	Identity id.Identity
	Name     string
	Avatar   object.ID
}

type UserDesc struct {
	Name string
}

func (UserDesc) Type() string {
	return "user"
}
