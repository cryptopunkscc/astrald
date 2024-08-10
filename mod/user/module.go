package user

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "user"
const DBPrefix = "users__"

type Module interface {
	UserID() *astral.Identity
	SetUserID(userID *astral.Identity) error

	Nodes(userID *astral.Identity) []*astral.Identity
	Owner(nodeID *astral.Identity) *astral.Identity
}
