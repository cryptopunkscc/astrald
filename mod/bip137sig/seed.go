package bip137sig

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Object = Seed{}

type Seed struct {
	Data []byte
}

func (s Seed) ObjectType() string {
	return "bip137sig.seed"
}

func (s Seed) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&s).WriteTo(w)
}

func (s Seed) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(s).ReadFrom(r)
}

func init() {
	_ = astral.Add(&Seed{})
}
