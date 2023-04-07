package peers

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"sync"
)

type PeerSet struct {
	peers   []*Peer
	idIndex map[string]struct{}
	mu      sync.Mutex
}

func NewPeerSet() *PeerSet {
	return &PeerSet{
		peers:   make([]*Peer, 0),
		idIndex: make(map[string]struct{}),
	}
}

// Add adds a peer to the set
func (set *PeerSet) Add(peer *Peer) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	var pkh = peer.Identity().PublicKeyHex()
	if _, found := set.idIndex[pkh]; found {
		return errors.New("peer already added to set")
	}

	set.peers = append(set.peers, peer)
	set.idIndex[pkh] = struct{}{}

	return nil
}

// Find returns a peer from the set by its identity, nil if peer is not in the set.
func (set *PeerSet) Find(id id.Identity) *Peer {
	set.mu.Lock()
	defer set.mu.Unlock()

	var keyHex = id.PublicKeyHex()

	if _, found := set.idIndex[keyHex]; !found {
		return nil
	}

	for _, peer := range set.peers {
		if peer.Identity().IsEqual(id) {
			return peer
		}
	}

	return nil
}

// Contains returns true if the set contains a peer with this identity
func (set *PeerSet) Contains(id id.Identity) (found bool) {
	set.mu.Lock()
	defer set.mu.Unlock()

	_, found = set.idIndex[id.PublicKeyHex()]
	return
}

// Remove removes a peer from the set
func (set *PeerSet) Remove(peer *Peer) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	var keyHex = peer.Identity().PublicKeyHex()

	if _, found := set.idIndex[keyHex]; !found {
		return errors.New("peer not in set")
	}

	idx := -1
	for i, p := range set.peers {
		if peer == p {
			idx = i
			break
		}
	}

	set.peers = append(set.peers[:idx], set.peers[idx+1:]...)
	delete(set.idIndex, keyHex)

	return nil
}

// All returns a copy of the list of all peers in the set.
func (set *PeerSet) All() []*Peer {
	set.mu.Lock()
	defer set.mu.Unlock()

	var all = make([]*Peer, len(set.peers))
	copy(all, set.peers)

	return all
}

// Count returns the number of peers in the set
func (set *PeerSet) Count() int {
	return len(set.peers)
}
