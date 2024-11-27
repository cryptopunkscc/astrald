package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"time"
)

const ModuleName = "user"
const DBPrefix = "users__"

const ActionClaim = "astrald.mod.user.claim"

type Module interface {
	UserID() *astral.Identity
	SetUserID(userID *astral.Identity) error

	Nodes(userID *astral.Identity) []*astral.Identity
	Owner(nodeID *astral.Identity) *astral.Identity

	Remote(targetID *astral.Identity, callerID *astral.Identity) (Consumer, error)
}

type Consumer interface {
	Claim(context.Context, time.Duration) (err error)
}
