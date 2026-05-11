package indexing

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Object = &ChangeAck{}

type ChangeAck struct {
	Repo    astral.String8
	Version astral.Uint64
}

func (ChangeAck) ObjectType() string { return "indexing.ack" }

func (a ChangeAck) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&a).WriteTo(w)
}

func (a *ChangeAck) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(a).ReadFrom(r)
}

func init() {
	_ = astral.Add(&ChangeAck{})
}
