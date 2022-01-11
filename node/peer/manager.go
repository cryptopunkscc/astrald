package peer

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
)

type Manager struct {
	peers map[string]*Peer
	mu    sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		peers: make(map[string]*Peer),
	}
}

func (m *Manager) Find(id id.Identity, create bool) *Peer {
	m.mu.Lock()
	defer m.mu.Unlock()

	hex := id.PublicKeyHex()
	if p, ok := m.peers[hex]; ok {
		return p
	}

	if create {
		m.peers[hex] = New(id)
	}

	return m.peers[hex]
}

func (m *Manager) All() <-chan *Peer {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan *Peer, len(m.peers))
	for _, p := range m.peers {
		ch <- p
	}
	close(ch)
	return ch
}

func (m *Manager) Query(ctx context.Context, remoteID id.Identity, query string) (*link.Conn, error) {
	peer := m.Find(remoteID, true)
	if peer == nil {
		return nil, errors.New("peer not linked")
	}

	if _, err := peer.WaitLinked(ctx); err != nil {
		return nil, err
	}

	return peer.Query(ctx, query)
}
