package astral

import (
	"io"
)

type Query struct {
	Nonce       Nonce
	Caller      *Identity
	Target      *Identity
	QueryString string
}

var _ Object = &Query{}

func (Query) ObjectType() string { return "query" }

// NewQuery returns a Query instance with a random Nonce.
func NewQuery(caller *Identity, target *Identity, query string) *Query {
	return &Query{
		Nonce:       NewNonce(),
		Caller:      caller,
		Target:      target,
		QueryString: query,
	}
}

// binary

func (s Query) WriteTo(w io.Writer) (int64, error) {
	return Objectify(&s).WriteTo(w)
}

func (s *Query) ReadFrom(r io.Reader) (int64, error) {
	return Objectify(s).ReadFrom(r)
}

// json

func (s Query) MarshalJSON() ([]byte, error) {
	return Objectify(&s).MarshalJSON()
}

func (s *Query) UnmarshalJSON(bytes []byte) error {
	return Objectify(s).UnmarshalJSON(bytes)
}

// ...

func init() {
	_ = Add(&Query{})
}
