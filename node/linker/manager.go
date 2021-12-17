package linker

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/peer"
	"sync"
)

type Manager struct {
	localID  id.Identity
	resolver contacts.Resolver
	dialer   infra.Dialer

	mu sync.Mutex

	// queues
	links    chan *link.Link
	linkerMu map[*peer.Peer]*sync.Mutex
}

func New(identity id.Identity, resolver contacts.Resolver, dialer infra.Dialer) *Manager {
	return &Manager{
		localID:  identity,
		resolver: resolver,
		dialer:   dialer,
		linkerMu: make(map[*peer.Peer]*sync.Mutex),
		links:    make(chan *link.Link, 1),
	}
}

func (m *Manager) Links() <-chan *link.Link {
	return m.links
}

func (m *Manager) NewLink(ctx context.Context, p *peer.Peer) {
	mu := m.peerMutex(p)
	mu.Lock()
	defer mu.Unlock()

	// prepare the contact resolver without already linked networks
	var r = m.resolver
	for _, name := range link.Networks(p.Links()) {
		r = contacts.Filter(r, contacts.SkipNetwork(name))
	}

	linker := ConcurrentLinker{
		LocalID:  m.localID,
		RemoteID: p.Identity(),
		Resolver: r,
		Dialer:   m.dialer,
	}

	link := linker.Link(ctx)
	if link != nil {
		m.links <- link
	}
}

func (m *Manager) Connect(ctx context.Context, p *peer.Peer) (*link.Link, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// see if we have a link already
	if l := link.Select(p.Links(), link.Fastest); l != nil {
		return l, nil
	}

	ch := make(chan *link.Link, 1)

	// wait for a link with the peer
	go func() {
		defer close(ch)
		links := p.FollowLinks(ctx, false)
		for {
			select {
			case <-ctx.Done():
				return

			case l := <-links:
				ch <- l
				return
			}
		}
	}()

	// try to produce a link using the default linker
	go m.NewLink(ctx, p)

	link, ok := <-ch
	if !ok {
		return nil, errors.New("peer unreachable")
	}

	return link, nil
}

func (m *Manager) peerMutex(p *peer.Peer) *sync.Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()

	if mu, ok := m.linkerMu[p]; ok {
		return mu
	}
	m.linkerMu[p] = &sync.Mutex{}
	return m.linkerMu[p]
}
