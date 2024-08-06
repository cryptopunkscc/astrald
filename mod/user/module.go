package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

const ModuleName = "user"
const DBPrefix = "users__"

type Module interface {
	UserID() *astral.Identity
	SetUserID(userID *astral.Identity) error

	Nodes(userID *astral.Identity) []*astral.Identity
	Owner(nodeID *astral.Identity) *astral.Identity
}

type Profile struct {
	Identity *astral.Identity
	Name     string
	Avatar   object.ID
}

type UserDesc struct {
	Name string
}

func (UserDesc) Type() string {
	return "user"
}
