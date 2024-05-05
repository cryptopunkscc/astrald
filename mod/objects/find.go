package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/object"
)

type Finder interface {
	Find(ctx context.Context, query string, opts *FindOpts) ([]Match, error)
}

type FindOpts struct {
	Zone   Zone
	Filter id.Filter
}

type Match struct {
	ObjectID object.ID
	Score    int
	Exp      string
}
