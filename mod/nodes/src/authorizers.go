package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

func (mod *Module) AuthorizeRelayFor(ctx *astral.Context, a *nodes.RelayForAction) bool {
	return a.Actor().IsEqual(a.ForID)
}
