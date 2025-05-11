package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

func (mod *Module) Authorize(identity *astral.Identity, action auth.Action, target astral.Object) bool {
	switch action {
	case nodes.ActionRelayFor:
		appID, ok := target.(*astral.Identity)
		if !ok {
			return false
		}

		c, _ := mod.db.FindActiveAppContractsByAppAndHost(appID, identity)
		if len(c) > 0 {
			return true
		}
	}
	
	return false
}
