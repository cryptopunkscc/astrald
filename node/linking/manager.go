package linking

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/link"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/peer"
	"sync"
	"time"
)

type Manager struct {
	localID id.Identity

	contacts *contacts.Manager
	peers    *peer.Manager
	dialer   infra.Dialer

	mu         sync.Mutex
	optimizers map[string]*optimizer
	context    context.Context
	links      chan *link.Link
}

func Run(ctx context.Context, localID id.Identity, contacts *contacts.Manager, peers *peer.Manager, dialer infra.Dialer) *Manager {
	return &Manager{
		localID:    localID,
		contacts:   contacts,
		peers:      peers,
		dialer:     dialer,
		context:    ctx,
		optimizers: make(map[string]*optimizer),
		links:      make(chan *link.Link, 1),
	}
}

func (m *Manager) Optimize(remoteID id.Identity, d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hexID := remoteID.PublicKeyHex()
	if opt, found := m.optimizers[hexID]; found {
		opt.optimize(d)
		return
	}

	peerOpt := NewPeerOptimizer(m.localID, remoteID, m.contacts, m.peers, m.dialer, func(l *link.Link) error {
		m.links <- l
		return nil
	})

	opt := newOptimizer(peerOpt, d)
	m.optimizers[hexID] = opt
	opt.Start()

	//TODO: Emit an event for logging?
	//log.Println("linking", m.contacts.DisplayName(remoteID))

	go func() {
		opt.wait(m.context)
		m.removeOptimizer(remoteID)
	}()
}

func (m *Manager) Links() chan *link.Link {
	return m.links
}

func (m *Manager) removeOptimizer(remoteID id.Identity) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hexID := remoteID.PublicKeyHex()

	if opt, found := m.optimizers[hexID]; found {
		opt.Stop()
		delete(m.optimizers, hexID)

		//TODO: Emit an event for logging?
		//log.Println("stopped linking", m.contacts.DisplayName(remoteID))
	}
}
