package peer

import (
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"sync"
)

type Pool struct {
	peers  map[string]*Peer
	peerMu sync.Mutex
}

func NewPool() *Pool {
	return &Pool{
		peers: make(map[string]*Peer),
	}
}

func (pool *Pool) Peer(peerID id.Identity) *Peer {
	pool.peerMu.Lock()
	defer pool.peerMu.Unlock()

	hex := peerID.PublicKeyHex()
	if peer, found := pool.peers[hex]; found {
		return peer
	}
	pool.peers[hex] = New(peerID)
	return pool.peers[hex]
}

func (pool *Pool) All() <-chan *Peer {
	pool.peerMu.Lock()
	defer pool.peerMu.Unlock()

	ch := make(chan *Peer, len(pool.peers))
	for _, p := range pool.peers {
		ch <- p
	}
	close(ch)
	return ch
}

func (pool *Pool) Count() int {
	return len(pool.peers)
}

func (pool *Pool) AddLink(link *link.Link) error {
	return pool.Peer(link.RemoteIdentity()).AddLink(link)
}
