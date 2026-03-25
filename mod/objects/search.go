package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
)

// Searcher is used to search for objects matching a query
type Searcher interface {
	SearchObject(ctx *astral.Context, query SearchQuery) (<-chan *SearchResult, error)
}

type SearchPreprocessor interface {
	PreprocessSearch(*Search)
}

type Search struct {
	CallerID *astral.Identity
	Query    SearchQuery
	Sources  []*astral.Identity
}
