package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

func (mod *Module) AuthorizeNodesRelayFor(ctx *astral.Context, action *nodes.RelayForAction) bool {
	ctx = ctx.ExcludeZone(astral.ZoneNetwork)
	contracts, err := mod.Auth.SignedContracts().
		WithIssuer(action.CallerID).
		WithSubject(action.Actor()).
		WithAction(&apphost.HostForAction{}).
		Find(ctx)
	return err == nil && len(contracts) > 0
}
