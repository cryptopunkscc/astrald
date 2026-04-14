package nodes

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

// RelayForAction requests permission for ActorId to relay traffic on behalf of CallerID.
type RelayForAction struct {
	auth.Action
	CallerID *astral.Identity
}

func (RelayForAction) ObjectType() string { return "mod.nodes.relay_for_action" }

func (a RelayForAction) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&a).WriteTo(w)
}

func (a *RelayForAction) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(a).ReadFrom(r)
}

func (a RelayForAction) ApplyConstraints(cs *astral.Bundle) bool {
	return cs == nil || len(cs.Objects()) == 0
}

func init() { _ = astral.Add(&RelayForAction{}) }
