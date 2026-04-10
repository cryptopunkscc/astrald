package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

// Authorize dispatches the action to all registered local handlers.
// Returns true on the first handler that allows; false if none match.
func (mod *Module) Authorize(ctx *astral.Context, action auth.ActionObject) bool {
	actionType := action.ObjectType()
	actor := action.Actor()

	for _, h := range mod.get(actionType) {
		if h.Authorize(ctx, action) {
			mod.log.Logv(1, "allow %v %v", actor, actionType)
			return true
		}
	}

	if mod.authorizeViaContracts(ctx, actor, actionType, action) {
		mod.log.Logv(1, "allow %v %v via contract", actor, actionType)
		return true
	}

	mod.log.Logv(2, "deny %v %v", actor, actionType)
	return false
}

func (mod *Module) authorizeViaContracts(ctx *astral.Context, actorID *astral.Identity, actionType string, action auth.ActionObject) bool {
	contracts, err := mod.db.findActiveContractsBySubject(actorID)
	if err != nil {
		mod.log.Logv(1, "error finding active contracts: %v", err)
		return false
	}

	for _, contract := range contracts {
		sc, err := objects.Load[*auth.SignedContract](ctx, mod.Objects.ReadDefault(), contract.ObjectID)
		if err != nil {
			continue
		}

		if err := mod.verifySignedContract(sc); err != nil {
			continue
		}

		for _, permit := range sc.Permits {
			if string(permit.Action) != actionType {
				continue
			}

			if c, ok := action.(auth.Constrainable); ok {
				if !c.ApplyConstraints(permit.Constraints) {
					continue
				}
			}

			return true
		}
	}
	return false
}
