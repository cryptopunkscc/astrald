package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

// Authorize authorizes node's identity to perform all actions
func (mod *Module) Authorize(id id.Identity, action string, target astral.Object) bool {
	switch action {
	case objects.ActionRead,
		objects.ActionWrite,
		objects.ActionPurge:
	default:
		return false
	}

	if id.IsEqual(mod.node.Identity()) {
		return true
	}

	return false
}
