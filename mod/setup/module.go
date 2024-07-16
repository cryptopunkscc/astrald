package setup

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
)

const ModuleName = "setup"

type Module interface {
	Invite(ctx context.Context, userID id.Identity, nodeID id.Identity) error
}
