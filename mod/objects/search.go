package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
)

// Searcher is used to search for objects matching a query
type Searcher interface {
	SearchObject(ctx *astral.Context, query string, opts *SearchOpts) (<-chan *SearchResult, error)
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

// SearchArgs contains arguments to the objects.search call
type SearchArgs struct {
	Query string      `query:"key:q"`
	Zone  astral.Zone `query:"optional"`
	Out   string      `query:"optional"`
	Ext   string      `query:"optional"`
}
