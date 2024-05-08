package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/object"
)

type Finder interface {
	Find(ctx context.Context, query string, opts *FindOpts) ([]Match, error)
}

type FindOpts struct {
	net.Scope
}

type Match struct {
	ObjectID object.ID
	Score    int
	Exp      string
}

func DefaultFindOpts() *FindOpts {
	return &FindOpts{
		Scope: net.DefaultScope(),
	}
}
