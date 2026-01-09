package user

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Object = &Revoker{}

type Revoker struct {
	ID  *astral.Identity
	Sig astral.Bytes8
}

func (s Revoker) ObjectType() string {
	return "mod.users.revoker"
}

func (s Revoker) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(s).WriteTo(w)
}

func (s Revoker) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(s).ReadFrom(r)
}

func init() {
	_ = astral.DefaultBlueprints.Add(&Revoker{})
}
