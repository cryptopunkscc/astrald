package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/network/contacts"
)

const defaultConcurrency = 8

// ConcurrentLinker tries to establish a link on any address from the resolver using several concurrent dialers.
type ConcurrentLinker struct {
	LocalID     id.Identity
	RemoteID    id.Identity
	Resolver    contacts.Resolver
	Concurrency int
}

func (l *ConcurrentLinker) getConcurrency() int {
	if l.Concurrency == 0 {
		return defaultConcurrency
	}
	return l.Concurrency
}

func (l *ConcurrentLinker) Link(ctx context.Context) *link.Link {
	// get current addresses for the node
	addrs := l.Resolver.Resolve(l.RemoteID)

	// try to link
	rawLink := astral.LinkFirst(ctx,
		l.LocalID,
		l.RemoteID,
		astral.DialMany(ctx,
			addrs,
			l.getConcurrency(),
		),
	)

	if rawLink == nil {
		return nil
	}

	return link.Wrap(rawLink)
}
