package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

// Authorize dispatches the action to all registered local handlers.
// Returns true on the first handler that allows; false if none match.
func (mod *Module) Authorize(ctx *astral.Context, action auth.ActionObject) bool {
	actionType := action.ObjectType()
	actor := action.Actor()

	if mod.authorizeHandlers(ctx, action) {
		mod.log.Logv(1, "allow %v %v", actor, actionType)
		return true
	}

	contracts, err := mod.SignedContracts().WithSubject(actor).WithAction(action).Find(ctx)
	if err != nil {
		mod.log.Logv(1, "error finding active contracts: %v", err)
		return false
	}

	for _, sc := range contracts {
		// todo: find better way to change actor of an action before running handlers
		action.SetActor(sc.Issuer)
		allowed := mod.authorizeHandlers(ctx, action)
		if allowed {
			mod.log.Logv(1, "allow %v %v", actor, actionType)
			return true
		}
	}

	return false
}

func (mod *Module) authorizeHandlers(ctx *astral.Context, action auth.ActionObject) bool {
	for _, h := range mod.get(action.ObjectType()) {
		if h.Authorize(ctx, action) {
			return true
		}
	}
	return false
}
