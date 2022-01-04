package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
)

type Manager struct {
	linker Linker
	mu     sync.Mutex
	links  chan *link.Link
	peers  map[string]*sync.Mutex
}

func New(identity id.Identity, resolver contacts.Resolver, dialer infra.Dialer) *Manager {
	return &Manager{
		linker: &ConcurrentLinker{
			LocalID:  identity,
			Resolver: resolver,
			Dialer:   dialer,
		},

		peers: make(map[string]*sync.Mutex),
		links: make(chan *link.Link, 1),
	}
}

func (m *Manager) Links() <-chan *link.Link {
	return m.links
}

func (m *Manager) Link(ctx context.Context, identity id.Identity) {
	mu := m.peerMutex(identity)

	mu.Lock()
	defer mu.Unlock()

	link := m.linker.Link(ctx, identity)
	if link != nil {
		m.links <- link
	}
}

func (m *Manager) peerMutex(identity id.Identity) *sync.Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()

	hex := identity.PublicKeyHex()
	if mu, ok := m.peers[hex]; ok {
		return mu
	}
	m.peers[hex] = &sync.Mutex{}
	return m.peers[hex]
}
