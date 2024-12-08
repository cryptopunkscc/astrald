package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

var _ astral.Object = &SearchResult{}

type SearchResult struct {
	ObjectID object.ID
}

func (*SearchResult) ObjectType() string { return "astrald.mod.objects.search_result" }

func (sr SearchResult) WriteTo(w io.Writer) (n int64, err error) {
	return sr.ObjectID.WriteTo(w)
}

func (sr *SearchResult) ReadFrom(r io.Reader) (n int64, err error) {
	return sr.ObjectID.ReadFrom(r)
}

func (sr SearchResult) String() string {
	return sr.ObjectID.String()
}
