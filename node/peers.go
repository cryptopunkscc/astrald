package node

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/peer"
)

func (node *Node) Peer(id id.Identity) *peer.Peer {
	node.peersMu.Lock()
	defer node.peersMu.Unlock()

	return node.peers[id.PublicKeyHex()]
}

func (node *Node) Peers() <-chan *peer.Peer {
	node.peersMu.Lock()
	defer node.peersMu.Unlock()

	ch := make(chan *peer.Peer, len(node.peers))
	for _, p := range node.peers {
		ch <- p
	}
	close(ch)
	return ch
}

func (node *Node) makePeer(id id.Identity) *peer.Peer {
	node.peersMu.Lock()
	defer node.peersMu.Unlock()

	hex := id.PublicKeyHex()
	if p, ok := node.peers[hex]; ok {
		return p
	}
	node.peers[hex] = peer.New(id)
	return node.peers[hex]
}
