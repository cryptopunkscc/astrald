package core

import "github.com/cryptopunkscc/astrald/app/warpdrive/api"

type peerManager core

func (c *core) Peer() api.PeerManager {
	return (*peerManager)(c)
}

func (c *peerManager) Update(peerId string, attr string, value string) {
	id := api.PeerId(peerId)
	peer := c.peers[id]
	cached := peer != nil
	if !cached {
		peer = &api.Peer{Id: id}
		c.peers[id] = peer
	}
	switch attr {
	case "mod":
		peer.Mod = value
	case "alias":
		peer.Alias = value
	default:
		if cached {
			return
		}
	}
	var peers []api.Peer
	for _, p := range c.peers {
		peers = append(peers, *p)
	}
	c.Peers().Save(peers)
}

func (c *peerManager) Get(id api.PeerId) api.Peer {
	peer := c.peers[id]
	if peer == nil {
		peer = &api.Peer{
			Id:    id,
			Alias: "",
			Mod:   "",
		}
		c.peers[id] = peer
	}
	return *peer
}

func (c *peerManager) List() (peers []*api.Peer) {
	peers = make([]*api.Peer, len(c.peers))
	i := 0
	peersMap := c.peers
	for key := range peersMap {
		peers[i] = c.peers[key]
		i++
	}
	return
}

func (c *peerManager) Offers() *api.Subscriptions {
	return c.filesOffers
}
