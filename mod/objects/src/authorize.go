package objects

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

// Authorize authorizes node's identity to perform all actions
func (mod *Module) Authorize(id id.Identity, action string, args ...any) bool {
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

	if len(args) == 0 {
		return false
	}
	objectID, ok := args[0].(object.ID)
	if !ok {
		return false
	}

	return mod.isHolding(id, objectID)
}
