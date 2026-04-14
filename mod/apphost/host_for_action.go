package apphost

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

type HostForAction struct {
	auth.Action
}

func (HostForAction) ObjectType() string { return "mod.apphost.host_for_action" }

func (a HostForAction) WriteTo(w io.Writer) (int64, error)   { return astral.Objectify(&a).WriteTo(w) }
func (a *HostForAction) ReadFrom(r io.Reader) (int64, error) { return astral.Objectify(a).ReadFrom(r) }

func init() { _ = astral.Add(&HostForAction{}) }
