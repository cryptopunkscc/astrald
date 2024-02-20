package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
)

type Finder interface {
	Find(ctx context.Context, query string, opts *FindOpts) ([]Match, error)
}

type FindOpts struct {
	Network bool
	Virtual bool
	Filter  id.Filter
}

type Match struct {
	DataID data.ID
	Score  int
	Exp    string
}
