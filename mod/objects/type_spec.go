package objects

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
)

type TypeSpec struct {
	Name   string
	Fields []query.FieldSpec
}

var _ astral.Object = &TypeSpec{}

func (TypeSpec) ObjectType() string {
	return "objects.type_spec"
}

// binary

func (s TypeSpec) WriteTo(w io.Writer) (int64, error) {
	return astral.Objectify(&s).WriteTo(w)
}

func (s *TypeSpec) ReadFrom(r io.Reader) (int64, error) {
	return astral.Objectify(s).ReadFrom(r)
}

// json

func (s TypeSpec) MarshalJSON() ([]byte, error) {
	return astral.Objectify(&s).MarshalJSON()
}

func (s *TypeSpec) UnmarshalJSON(bytes []byte) error {
	return astral.Objectify(s).UnmarshalJSON(bytes)
}

// ...

func init() {
	_ = astral.Add(&TypeSpec{})
}
