package objects

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

type SearchResult struct {
	ObjectID *object.ID
}

// astral

var _ astral.Object = &SearchResult{}

func (SearchResult) ObjectType() string { return "astrald.mod.objects.search_result" }

func (sr SearchResult) WriteTo(w io.Writer) (n int64, err error) {
	return sr.ObjectID.WriteTo(w)
}

func (sr *SearchResult) ReadFrom(r io.Reader) (n int64, err error) {
	return sr.ObjectID.ReadFrom(r)
}

// json

func (sr SearchResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(sr.ObjectID)
}

func (sr *SearchResult) UnmarshalJSON(bytes []byte) error {
	sr.ObjectID = &object.ID{}
	return json.Unmarshal(bytes, sr.ObjectID)
}

// text

func (sr SearchResult) MarshalText() (text []byte, err error) {
	return sr.ObjectID.MarshalText()
}

func (sr *SearchResult) UnmarshalText(text []byte) error {
	sr.ObjectID = &object.ID{}
	return sr.ObjectID.UnmarshalText(text)
}

// other

func (sr SearchResult) String() string {
	return sr.ObjectID.String()
}
