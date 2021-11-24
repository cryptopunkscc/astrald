package peer

import (
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"sync"
)

type Set struct {
	peers map[string]*Peer
	mu    sync.Mutex
}

func NewSet() *Set {
	return &Set{
		peers: make(map[string]*Peer),
	}
}

func (set *Set) Peer(peerID id.Identity) *Peer {
	set.mu.Lock()
	defer set.mu.Unlock()

	strID := peerID.PublicKeyHex()
	if peer, found := set.peers[strID]; found {
		return peer
	}
	set.peers[strID] = New(peerID)
	return set.peers[strID]
}

func (set *Set) Each() <-chan *Peer {
	set.mu.Lock()
	defer set.mu.Unlock()

	ch := make(chan *Peer, len(set.peers))
	for _, p := range set.peers {
		ch <- p
	}
	close(ch)
	return ch
}

func (set *Set) AddLink(link *link.Link) error {
	return set.Peer(link.RemoteIdentity()).Add(link)
}
