package service

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/file"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/memory"
)

var _ api.PeerService = Peer{}

type Peer api.Core

func (c Peer) Update(peerId string, attr string, value string) {
	c.Mutex.Peers.Lock()
	defer c.Mutex.Peers.Unlock()
	id := api.PeerId(peerId)
	mem := memory.Peers(c).Get()
	peer := mem[id]
	cached := peer != nil
	if !cached {
		peer = &api.Peer{Id: id}
		mem[id] = peer
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
	for _, p := range mem {
		peers = append(peers, *p)
	}
	file.Peers(c).Save(peers)
}

func (c Peer) Get(id api.PeerId) api.Peer {
	c.Mutex.Peers.Lock()
	defer c.Mutex.Peers.Unlock()
	mem := memory.Peers(c).Get()
	peer := mem[id]
	if peer == nil {
		peer = &api.Peer{
			Id:    id,
			Alias: "",
			Mod:   "",
		}
		mem[id] = peer
	}
	return *peer
}

func (c Peer) List() (peers []api.Peer) {
	c.Mutex.Peers.Lock()
	defer c.Mutex.Peers.Unlock()
	return memory.Peers(c).List()
}

func (c Peer) Offers() *api.Subscriptions {
	return c.FilesOffers
}
