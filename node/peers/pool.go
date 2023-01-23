package peers

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
)

type Pool struct {
	id      id.Identity
	peers   map[string]*Peer
	queue   *sig.Queue
	mu      sync.Mutex
	queries chan *link.Query
	events  event.Queue
}

func NewPool(localID id.Identity, eventParent *event.Queue) *Pool {
	p := &Pool{
		id:      localID,
		peers:   make(map[string]*Peer),
		queries: make(chan *link.Query),
		queue:   &sig.Queue{},
	}

	p.events.SetParent(eventParent)

	return p
}

func (pool *Pool) Queries() <-chan *link.Query {
	return pool.queries
}

func (pool *Pool) Peer(remoteID id.Identity) *Peer {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.peer(remoteID)
}

func (pool *Pool) Peers(follow context.Context) <-chan *Peer {
	pool.mu.Lock()
	ch := make(chan *Peer, len(pool.peers))
	for _, p := range pool.peers {
		ch <- p
	}
	pool.mu.Unlock()

	if follow == nil {
		close(ch)
		return ch
	}

	newPeers := pool.queue.Subscribe(follow)
	go func() {
		defer close(ch)
		for p := range newPeers {
			ch <- p.(*Peer)
		}
	}()
	return ch
}

func (pool *Pool) peer(remoteID id.Identity) *Peer {
	if peer, found := pool.peers[remoteID.String()]; found {
		return peer
	}
	return nil
}

func (pool *Pool) addLink(l *link.Link) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if !l.LocalIdentity().IsEqual(pool.id) {
		return errors.New("local identity mismatch")
	}

	peer, isNew := pool.createPeer(l.RemoteIdentity())

	if err := peer.addLink(l); err != nil {
		if isNew {
			pool.deletePeer(peer.Identity())
		}
		return err
	}

	if isNew {
		// forward peer queries
		go func() {
			for q := range peer.Queries() {
				pool.queries <- q
			}
		}()

		pool.queue = pool.queue.Push(peer)
		log.Info("%s linked", peer.Identity())
		pool.events.Emit(EventPeerLinked{Peer: peer, Link: l})
	}

	return nil
}

func (pool *Pool) removeLink(l *link.Link) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	peer := pool.peer(l.RemoteIdentity())
	if peer == nil {
		return errors.New("not found")
	}

	if err := peer.removeLink(l); err != nil {
		return err
	}

	if peer.LinkCount() == 0 {
		pool.deletePeer(peer.Identity())
		log.Info("%s unlinked", peer.Identity())
		pool.events.Emit(EventPeerUnlinked{Peer: peer})
	}

	return nil
}

func (pool *Pool) createPeer(nodeID id.Identity) (*Peer, bool) {
	hexID := nodeID.PublicKeyHex()

	if peer, found := pool.peers[hexID]; found {
		return peer, false
	}

	pool.peers[hexID] = NewPeer(nodeID, &pool.events)

	return pool.peers[hexID], true
}

func (pool *Pool) deletePeer(nodeID id.Identity) {
	delete(pool.peers, nodeID.PublicKeyHex())
}
