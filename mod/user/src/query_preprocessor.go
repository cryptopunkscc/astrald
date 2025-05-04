package user

import (
	"github.com/cryptopunkscc/astrald/core"
)

var _ core.QueryPreprocessor = &Module{}

func (mod *Module) PreprocessQuery(qm *core.QueryModifier) error {
	// get the active contract
	ac := mod.ActiveContract()
	if ac == nil {
		return nil
	}

	// if the query is coming from the active user, attach the active contract
	if qm.Query().Caller.IsEqual(ac.UserID) {
		qm.Attach(ac)
	}

	if qm.Query().Target.IsEqual(ac.UserID) {
		// if the target is the active user, attach all siblings
		for _, sib := range mod.getLinkedSibs() {
			qm.AddRelay(sib)
		}
	} else {
		// ... otherwise attach all active nodes
		for _, relay := range mod.ActiveNodes(qm.Query().Target) {
			qm.AddRelay(relay)
		}
	}

	return nil
}
