package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
)

func (mod *Module) AuthorizeNodesRelayFor(_ *astral.Context, identity *astral.Identity, appID *astral.Identity) bool {
	c, _ := mod.db.FindActiveAppContractsByAppAndHost(appID, identity)
	return len(c) > 0
}
