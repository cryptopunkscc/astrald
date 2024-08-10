package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

// Authorize authorizes node's identity to perform all actions
func (mod *Module) Authorize(id *astral.Identity, action string, target astral.Object) bool {
	switch action {
	case objects.ActionRead,
		objects.ActionWrite,
		objects.ActionSearch,
		objects.ActionAccessDescriptor,
		objects.ActionPurge:
	default:
		return false
	}

	return id.IsEqual(mod.node.Identity())
}
