package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

// Searcher is used to find objects matching a query
type Searcher interface {
	Search(ctx context.Context, query string, opts *SearchOpts) ([]Match, error)
}

type SearchOpts struct {
	*astral.Scope
}

type Match struct {
	ObjectID object.ID
	Score    int
	Exp      string
}

func DefaultSearchOpts() *SearchOpts {
	return &SearchOpts{
		Scope: astral.DefaultScope(),
	}
}
