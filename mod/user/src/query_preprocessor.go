package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

var _ core.QueryPreprocessor = &Module{}

func (mod *Module) PreprocessQuery(q *astral.Query) error {
	mod.attachCallerProof(q)
	mod.attachRelays(q)

	return nil
}

func (mod *Module) attachCallerProof(q *astral.Query) {
	proof := mod.ActiveContract()
	if proof == nil {
		return
	}

	if !q.Caller.IsEqual(proof.UserID) {
		return
	}

	q.Extra.Set(nodes.ExtraCallerProof, proof)
}

func (mod *Module) attachRelays(q *astral.Query) {
	for _, relay := range mod.ActiveNodes(q.Target) {
		if relay.IsEqual(mod.node.Identity()) {
			continue
		}

		q.Extra.Set(nodes.ExtraRelayVia, relay)
	}
}
