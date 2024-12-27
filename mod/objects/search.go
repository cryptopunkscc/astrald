package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
)

// Searcher is used to search for objects matching a query
type Searcher interface {
	SearchObject(ctx context.Context, query string, opts *SearchOpts) (<-chan *SearchResult, error)
}

type SearchOpts struct {
	*astral.Scope
	ClientID *astral.Identity
	Extra    sig.Map[string, any]
}

func DefaultSearchOpts() *SearchOpts {
	return &SearchOpts{
		Scope: astral.DefaultScope(),
	}
}
