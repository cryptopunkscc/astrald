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

func (pool *Pool) AddLink(l *link.Link) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if !l.LocalIdentity().IsEqual(pool.id) {
		return errors.New("local identity mismatch")
	}

	remoteIDHex := l.RemoteIdentity().String()

	if peer, found := pool.peers[remoteIDHex]; found {
		return peer.AddLink(l)
	}

	peer := NewPeer(l.RemoteIdentity(), &pool.events)

	if err := peer.AddLink(l); err != nil {
		return err
	}

	pool.peers[remoteIDHex] = peer
	pool.queue = pool.queue.Push(peer)
	pool.events.Emit(EventLinked{Peer: peer, Link: l})

	go func() {
		for q := range peer.Queries() {
			pool.queries <- q
		}
		pool.mu.Lock()
		delete(pool.peers, remoteIDHex)
		pool.events.Emit(EventUnlinked{Peer: peer})
		pool.mu.Unlock()
	}()

	return nil
}

func (pool *Pool) Peer(remoteID id.Identity) *Peer {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if peer, found := pool.peers[remoteID.String()]; found {
		return peer
	}

	return nil
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

	newPeers := pool.queue.Subscribe(follow.Done())
	go func() {
		defer close(ch)
		for p := range newPeers {
			ch <- p.(*Peer)
		}
	}()
	return ch
}
