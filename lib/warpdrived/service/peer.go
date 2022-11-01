package service

import (
	"github.com/cryptopunkscc/astrald/lib/warpdrived/core"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/storage/file"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/storage/memory"
	"github.com/cryptopunkscc/astrald/proto/contacts"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
)

var _ warpdrive.PeerService = peer{}

type peer core.Component

func (srv peer) Fetch() {
	contactList, err := contacts.Client{Api: srv}.List()
	if err != nil {
		srv.Println("Cannot obtain contacts", err)
		return
	}
	srv.Mutex.Peers.Lock()
	defer srv.Mutex.Peers.Unlock()
	for _, contact := range contactList {
		srv.update(contact.Id, "alias", contact.Name)
	}
	srv.save()
}

func (srv peer) Update(peerId string, attr string, value string) {
	srv.Mutex.Peers.Lock()
	defer srv.Mutex.Peers.Unlock()
	srv.update(peerId, attr, value)
	srv.save()
}

func (srv peer) update(peerId string, attr string, value string) {
	id := warpdrive.PeerId(peerId)
	mem := memory.Peers(srv).Get()
	p := mem[id]
	cached := p != nil
	if !cached {
		p = &warpdrive.Peer{Id: id}
		mem[id] = p
	}
	switch attr {
	case "mod":
		p.Mod = value
	case "alias":
		p.Alias = value
	default:
		if cached {
			return
		}
	}
}

func (srv peer) save() {
	var peers []warpdrive.Peer
	mem := memory.Peers(srv).Get()
	for _, p := range mem {
		peers = append(peers, *p)
	}
	file.Peers(core.Component(srv)).Save(peers)
}

func (srv peer) Get(id warpdrive.PeerId) warpdrive.Peer {
	srv.Mutex.Peers.RLock()
	defer srv.Mutex.Peers.RUnlock()
	p := memory.Peers(srv).Get()[id]
	if p == nil {
		p = &warpdrive.Peer{
			Id:    id,
			Alias: "",
			Mod:   "",
		}
	}
	return *p
}

func (srv peer) List() (peers []warpdrive.Peer) {
	srv.Fetch()
	srv.Mutex.Peers.RLock()
	defer srv.Mutex.Peers.RUnlock()
	return memory.Peers(srv).List()
}

func (srv peer) Offers() *warpdrive.Subscriptions {
	return srv.IncomingOffers
}
