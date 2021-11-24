package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/network/route"
)

const defaultConcurrency = 8

// ConcurrentLinker tries to establish a link on any address from the router using several concurrent dialers.
type ConcurrentLinker struct {
	localID     id.Identity
	remoteID    id.Identity
	router      route.Router
	concurrency int
}

func (l *ConcurrentLinker) Concurrency() int {
	if l.concurrency == 0 {
		return defaultConcurrency
	}
	return l.concurrency
}

func (l *ConcurrentLinker) Link(ctx context.Context) *link.Link {
	// get current route for the node
	r := l.router.Route(l.remoteID)
	if r == nil {
		return nil
	}

	// try to link
	return astral.LinkFirst(ctx,
		l.localID,
		l.remoteID,
		astral.DialMany(ctx,
			r.Each(),
			l.Concurrency(),
		),
	)
}
