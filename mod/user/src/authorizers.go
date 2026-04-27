package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

// AuthorizeRelayFor allows a swarm node to relay queries on behalf of the local user.
func (mod *Module) AuthorizeRelayFor(ctx *astral.Context, a *nodes.RelayForAction) bool {
	if !a.ForID.IsEqual(mod.Identity()) {
		return false
	}
	for _, nodeID := range mod.LocalSwarm() {
		if nodeID.IsEqual(a.Actor()) {
			return true
		}
	}
	return false
}

func (mod *Module) AuthorizeReadObject(ctx *astral.Context, a *objects.ReadObjectAction) bool {
	if a.Actor().IsEqual(mod.Identity()) {
		return true
	}

	for _, nodeID := range mod.LocalSwarm() {
		if nodeID.IsEqual(a.Actor()) {
			return true
		}
	}

	return false
}
