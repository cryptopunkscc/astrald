package peer

import (
	_id "github.com/cryptopunkscc/astrald/node/auth/id"
	_link "github.com/cryptopunkscc/astrald/node/link"
	"sync"
)

// Peers keeps track of visibility and links to other nodes
type Peers struct {
	peers    map[string]*Peer
	mu       sync.Mutex
	requests chan _link.Request
}

func NewManager() *Peers {
	r := &Peers{
		peers:    make(map[string]*Peer),
		requests: make(chan _link.Request),
	}

	return r
}

func (peers *Peers) Requests() <-chan _link.Request {
	return peers.requests
}

func (peers *Peers) Peer(id _id.Identity) (*Peer, error) {
	peers.mu.Lock()
	defer peers.mu.Unlock()

	if peer, ok := peers.peers[id.String()]; ok {
		return peer, nil
	}

	peer := New(id)
	peers.peers[id.String()] = peer

	go func() {
		for req := range peer.Requests() {
			peers.requests <- req
		}
	}()

	return peer, nil
}

func (peers *Peers) AddLink(link *_link.Link) error {
	peer, _ := peers.Peer(link.RemoteIdentity())

	return peer.AddLink(link)
}
