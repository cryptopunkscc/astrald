package peer

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
)

type Manager struct {
	peers  map[string]*managedPeer
	mu     sync.Mutex
	events event.Queue
}

func NewManager(eventParent *event.Queue) *Manager {
	m := &Manager{
		peers: make(map[string]*managedPeer),
	}
	m.events.SetParent(eventParent)
	return m
}

func (m *Manager) Add(link *link.Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := sig.New()
	defer close(s)
	peer := m.hold(s, link.RemoteIdentity())

	if err := peer.peer.Add(link); err != nil {
		return err
	}

	peer.hold(link.Wait())

	return nil
}

func (m *Manager) Find(id id.Identity) *Peer {
	m.mu.Lock()
	defer m.mu.Unlock()

	hex := id.PublicKeyHex()
	if p, ok := m.peers[hex]; ok {
		return p.peer
	}
	return nil
}

func (m *Manager) Subscribe(cancel sig.Signal) <-chan event.Event {
	return m.events.Subscribe(cancel)
}

// Hold returns a peer and makes sure the maneger will not remove it until the context is done
func (m *Manager) Hold(ctx context.Context, identity id.Identity) *Peer {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.hold(ctx.Done(), identity).peer
}

func (m *Manager) hold(done sig.Signal, identity id.Identity) *managedPeer {
	hex := identity.PublicKeyHex()
	peer, ok := m.peers[hex]
	if ok {
		peer.hold(done)
		return peer
	}

	return m.create(done, identity)
}

func (m *Manager) create(done sig.Signal, identity id.Identity) *managedPeer {
	var peer = &managedPeer{peer: New(identity)}

	m.peers[identity.PublicKeyHex()] = peer
	peer.hold(done)

	peer.peer.events.SetParent(&m.events)

	go func() {
		peer.wg.Wait()
		m.remove(identity)
	}()

	return peer
}

func (m *Manager) remove(identity id.Identity) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	hex := identity.PublicKeyHex()
	if _, found := m.peers[hex]; !found {
		return errors.New("identity not found")
	}

	delete(m.peers, hex)

	return nil
}

func (m *Manager) All() <-chan *Peer {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan *Peer, len(m.peers))
	for _, p := range m.peers {
		ch <- p.peer
	}
	close(ch)
	return ch
}

func (m *Manager) Query(ctx context.Context, remoteID id.Identity, query string) (*link.Conn, error) {
	peer := m.Hold(ctx, remoteID)
	if peer == nil {
		return nil, errors.New("peer not linked")
	}

	if _, err := peer.WaitLinked(ctx); err != nil {
		return nil, err
	}

	return peer.Query(ctx, query)
}
