package apphost

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// HandleQueryMsg represents a request to handle a query.
type HandleQueryMsg struct {
	AuthToken astral.Nonce
	ID        astral.Nonce
	Caller    *astral.Identity
	Target    *astral.Identity
	Query     astral.String16
}

func (HandleQueryMsg) ObjectType() string { return "mod.apphost.handle_query_msg" }

func (msg HandleQueryMsg) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&msg).WriteTo(w)
}

func (msg *HandleQueryMsg) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(msg).ReadFrom(r)
}

func init() {
	_ = astral.DefaultBlueprints.Add(&HandleQueryMsg{})
}
