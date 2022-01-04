package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/link"
)

const DefaultConcurrency = 8

// ConcurrentLinker tries to establish a link on any address from the resolver using several concurrent dialers.
type ConcurrentLinker struct {
	LocalID     id.Identity
	Resolver    contacts.Resolver
	Dialer      infra.Dialer
	Concurrency int
}

func (linker *ConcurrentLinker) getConcurrency() int {
	if linker.Concurrency == 0 {
		return DefaultConcurrency
	}
	return linker.Concurrency
}

func (linker *ConcurrentLinker) Link(ctx context.Context, remoteID id.Identity) *link.Link {
	// get current addresses for the node
	contactAddr := linker.Resolver.Lookup(remoteID)

	// try to link
	rawLink := LinkFirst(ctx,
		linker.LocalID,
		remoteID,
		DialMany(ctx,
			linker.Dialer,
			extractInfraAddr(contactAddr),
			linker.getConcurrency(),
		),
	)

	if rawLink == nil {
		return nil
	}

	return link.Wrap(rawLink)
}

func extractInfraAddr(in <-chan *contacts.Addr) <-chan infra.Addr {
	out := make(chan infra.Addr)
	go func() {
		defer close(out)
		for i := range in {
			out <- i.Addr
		}
	}()
	return out
}
