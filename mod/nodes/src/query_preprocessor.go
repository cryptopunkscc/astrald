package nodes

import "github.com/cryptopunkscc/astrald/core"

func (mod *Module) PreprocessQuery(qm *core.QueryModifier) error {
	q := qm.Query()
	if r, ok := mod.relays.Delete(qm.Query().Nonce); ok {
		if !r.Target.IsZero() {
			q.Target = r.Target
		}
		if !r.Caller.IsZero() {
			q.Caller = r.Caller
		}
	}
	return nil
}
