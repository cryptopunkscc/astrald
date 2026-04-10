package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

func (mod *Module) AuthorizeNodesRelayFor(_ *astral.Context, action *nodes.RelayForAction) bool {
	c, _ := mod.db.FindActiveAppContractsByAppAndHost(action.CallerID, action.Actor())
	return len(c) > 0
}
