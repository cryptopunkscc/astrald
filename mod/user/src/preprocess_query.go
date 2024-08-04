package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

func (mod *Module) PreprocessQuery(q *astral.Query) error {
	mod.attachCallerProof(q)
	mod.attachRelays(q)

	return nil
}

func (mod *Module) attachCallerProof(q *astral.Query) {
	if !q.Caller.IsEqual(mod.userID) {
		return
	}

	proof, err := mod.LocalContract()
	if err != nil {
		return
	}

	q.Extra.Set(nodes.ExtraCallerProof, proof)
}

func (mod *Module) attachRelays(q *astral.Query) {
	for _, relay := range mod.Nodes(q.Target) {
		if relay.IsEqual(mod.node.Identity()) {
			continue
		}

		q.Extra.Set(nodes.ExtraRelayVia, relay)
	}
}
