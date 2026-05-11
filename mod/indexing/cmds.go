package indexing

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Object = &Index{}
var _ astral.Object = &Unindex{}

type Index struct {
	Repo     astral.String8
	Version  astral.Uint64
	ObjectID *astral.ObjectID
}

func (Index) ObjectType() string { return "indexing.index" }

func (i Index) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&i).WriteTo(w)
}

func (i *Index) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(i).ReadFrom(r)
}

type Unindex struct {
	Repo     astral.String8
	Version  astral.Uint64
	ObjectID *astral.ObjectID
}

func (Unindex) ObjectType() string { return "indexing.unindex" }

func (u Unindex) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&u).WriteTo(w)
}

func (u *Unindex) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(u).ReadFrom(r)
}

func init() {
	_ = astral.Add(&Index{})
	_ = astral.Add(&Unindex{})
}
