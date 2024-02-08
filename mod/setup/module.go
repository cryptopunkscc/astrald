package setup

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
)

const ModuleName = "setup"

type Module interface {
	Invite(ctx context.Context, userID id.Identity, nodeID id.Identity) error
}
