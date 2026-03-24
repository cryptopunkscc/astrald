package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

func (mod *Module) Authorize(identity *astral.Identity, action auth.Action, target astral.Object) bool {
	return auth.Auth(auth.ActionsMap{
		nodes.ActionRelayFor: {auth.NewHandler(mod.AuthorizeRelayFor)},
	}, identity, action, target)
}

func (mod *Module) AuthorizeRelayFor(identity *astral.Identity, appID *astral.Identity) bool {
	c, _ := mod.db.FindActiveAppContractsByAppAndHost(appID, identity)
	return len(c) > 0
}
