package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/network/graph"
)

const defaultConcurrency = 8

// ConcurrentLinker tries to establish a link on any address from the resolver using several concurrent dialers.
type ConcurrentLinker struct {
	localID     id.Identity
	remoteID    id.Identity
	resolver    graph.Resolver
	concurrency int
}

func (l *ConcurrentLinker) Concurrency() int {
	if l.concurrency == 0 {
		return defaultConcurrency
	}
	return l.concurrency
}

func (l *ConcurrentLinker) Link(ctx context.Context) *link.Link {
	// get current addresses for the node
	addrs := l.resolver.Resolve(l.remoteID)

	// try to link
	return astral.LinkFirst(ctx,
		l.localID,
		l.remoteID,
		astral.DialMany(ctx,
			addrs,
			l.Concurrency(),
		),
	)
}
