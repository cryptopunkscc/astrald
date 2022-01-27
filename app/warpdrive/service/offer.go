package service

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/file"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/memory"
	"sync"
)

var _ api.OfferService = Offer{}

type Offer struct {
	api.Core
	mu   *sync.Mutex
	subs *api.Subscriptions
	file api.OfferStorage
	mem  api.OfferStorage
}

func Outgoing(c api.Core) Offer {
	return Offer{
		Core: c,
		mu:   &c.Mutex.Outgoing,
		mem:  memory.Outgoing(c),
		file: file.Outgoing(c),
		subs: c.Observers.OutgoingStatus,
	}
}

func Incoming(c api.Core) Offer {
	return Offer{
		Core: c,
		mu:   &c.Mutex.Incoming,
		mem:  memory.Incoming(c),
		file: file.Incoming(c),
		subs: c.Observers.IncomingStatus,
	}
}

func (c Offer) Add(offerId string, files []api.Info, peer *api.Peer) {
	offer := &api.Offer{
		Files: files,
		Status: api.Status{
			Id: api.OfferId(offerId),
		},
	}
	if peer == nil {
		offer.Status.Status = "sent"
	} else {
		offer.Status.Status = "received"
		offer.Peer = peer.Id
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.mem.Save(offer)
	c.file.Save(offer)

	go c.notify(offer.Status, c.subs)

	if peer != nil && peer.Mod == api.PeerModAsk {
		go c.notify(offer, c.FilesOffers)
	}
}

func (c Offer) Update(offer *api.Offer, status string, persist bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	offer.Status.Status = status
	c.mem.Save(offer)
	if persist {
		c.file.Save(offer)
	}
	go c.notify(offer.Status, c.subs)
}

func (c Offer) Get() api.Offers {
	return c.mem.Get()
}

func (c Offer) Status() *api.Subscriptions {
	return c.subs
}

func (c Offer) notify(data interface{}, subscribers *api.Subscriptions) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		c.Println("Cannot create json from data", data, err)
		return
	}
	subscribers.Lock()
	defer subscribers.Unlock()
	for subscriber := range subscribers.Set {
		_, err := subscriber.Write(jsonData)
		if err != nil {
			c.Println("Error while sending files to subscriber", err)
			return
		}
	}
}
