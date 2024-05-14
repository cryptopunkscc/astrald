package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/object"
)

type Searcher interface {
	Search(ctx context.Context, query string, opts *SearchOpts) ([]Match, error)
}

type SearchOpts struct {
	net.Scope
}

type Match struct {
	ObjectID object.ID
	Score    int
	Exp      string
}

func DefaultSearchOpts() *SearchOpts {
	return &SearchOpts{
		Scope: net.DefaultScope(),
	}
}
