package bip137sig

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Object = Entropy{}

type Entropy struct {
	Data []byte
}

func (Entropy) ObjectType() string {
	return "bip137sig.entropy"
}

func (e Entropy) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&e).WriteTo(w)
}

func (e Entropy) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(e).ReadFrom(r)
}

func init() {
	_ = astral.Add(&Entropy{})
}
