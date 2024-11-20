package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

// Searcher is used to find objects matching a query
type Searcher interface {
	Search(ctx context.Context, query string, opts *SearchOpts) (<-chan *SearchResult, error)
}

type SearchOpts struct {
	*astral.Scope
}

type SearchResult struct {
	ObjectID object.ID
}

func (*SearchResult) ObjectType() string {
	return "astrald.mod.objects.search_result"
}

func (sr SearchResult) WriteTo(w io.Writer) (n int64, err error) {
	c := streams.NewWriteCounter(w)
	err = cslq.Encode(c, "v", sr.ObjectID)
	n = c.Total()
	return
}

func (sr *SearchResult) ReadFrom(r io.Reader) (n int64, err error) {
	c := streams.NewReadCounter(r)
	err = cslq.Decode(c, "v", &sr.ObjectID)
	n = c.Total()
	return
}

func (sr *SearchResult) String() string {
	return sr.ObjectID.String()
}

func DefaultSearchOpts() *SearchOpts {
	return &SearchOpts{
		Scope: astral.DefaultScope(),
	}
}
