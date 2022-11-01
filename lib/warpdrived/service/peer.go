package service

import (
	"github.com/cryptopunkscc/astrald/lib/warpdrived/core"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/storage/file"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/storage/memory"
	"github.com/cryptopunkscc/astrald/proto/contacts"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
)

var _ warpdrive.PeerService = Peer{}

type Peer core.Component

func (c Peer) Fetch() {
	contactList, err := contacts.Client{Api: c}.List()
	if err != nil {
		c.Println("Cannot obtain contacts", err)
		return
	}
	c.Mutex.Peers.Lock()
	defer c.Mutex.Peers.Unlock()
	for _, contact := range contactList {
		c.update(contact.Id, "alias", contact.Name)
	}
	c.save()
}

func (c Peer) Update(peerId string, attr string, value string) {
	c.Mutex.Peers.Lock()
	defer c.Mutex.Peers.Unlock()
	c.update(peerId, attr, value)
	c.save()
}

func (c Peer) update(peerId string, attr string, value string) {
	id := warpdrive.PeerId(peerId)
	mem := memory.Peers(c).Get()
	peer := mem[id]
	cached := peer != nil
	if !cached {
		peer = &warpdrive.Peer{Id: id}
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
}

func (c Peer) save() {
	var peers []warpdrive.Peer
	mem := memory.Peers(c).Get()
	for _, p := range mem {
		peers = append(peers, *p)
	}
	file.Peers(core.Component(c)).Save(peers)
}

func (c Peer) Get(id warpdrive.PeerId) warpdrive.Peer {
	c.Mutex.Peers.RLock()
	defer c.Mutex.Peers.RUnlock()
	peer := memory.Peers(c).Get()[id]
	if peer == nil {
		peer = &warpdrive.Peer{
			Id:    id,
			Alias: "",
			Mod:   "",
		}
	}
	return *peer
}

func (c Peer) List() (peers []warpdrive.Peer) {
	c.Fetch()
	c.Mutex.Peers.RLock()
	defer c.Mutex.Peers.RUnlock()
	return memory.Peers(c).List()
}

func (c Peer) Offers() *warpdrive.Subscriptions {
	return c.IncomingOffers
}
