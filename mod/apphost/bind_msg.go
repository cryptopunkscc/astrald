package apphost

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type BindMsg struct {
	Token astral.Nonce
}

func (BindMsg) ObjectType() string { return "mod.apphost.bind_msg" }

func (msg BindMsg) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&msg).WriteTo(w)
}

func (msg *BindMsg) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(msg).ReadFrom(r)
}

func init() {
	_ = astral.Add(&BindMsg{})
}
