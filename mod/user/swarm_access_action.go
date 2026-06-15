package user

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

// SwarmAccessAction is the permit type embedded in node contracts; its
// ObjectType string is what IsNodeContract checks to identify a valid swarm
// membership contract.
type SwarmAccessAction struct {
	auth.Action
}

func (SwarmAccessAction) ObjectType() string { return "mod.user.swarm_access_action" }

func (a SwarmAccessAction) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&a).WriteTo(w)
}

func (a *SwarmAccessAction) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(a).ReadFrom(r)
}

func init() { _ = astral.Add(&SwarmAccessAction{}) }
