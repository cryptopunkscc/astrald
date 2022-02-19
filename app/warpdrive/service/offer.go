package service

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/file"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/memory"
	"sort"
	"sync"
	"time"
)

type Offer struct {
	api.Core
	mu         *sync.RWMutex
	offerSubs  *api.Subscriptions
	statusSubs *api.Subscriptions
	file       api.OfferStorage
	mem        api.OfferStorage
	incoming   bool
}

func Outgoing(c api.Core) Offer {
	return Offer{
		Core:       c,
		mu:         &c.Mutex.Outgoing,
		mem:        memory.Outgoing(c),
		file:       file.Outgoing(c),
		offerSubs:  c.Observers.OutgoingOffers,
		statusSubs: c.Observers.OutgoingStatus,
	}
}

func Incoming(c api.Core) Offer {
	return Offer{
		Core:       c,
		mu:         &c.Mutex.Incoming,
		mem:        memory.Incoming(c),
		file:       file.Incoming(c),
		offerSubs:  c.Observers.IncomingOffers,
		statusSubs: c.Observers.IncomingStatus,
		incoming:   true,
	}
}

func (c Offer) Add(offerId string, files []api.Info, peerId api.PeerId) {
	createdAt := time.Now().UnixMilli()
	offer := &api.Offer{
		Files:  files,
		Peer:   peerId,
		Create: createdAt,
		Status: api.Status{
			Status: api.StatusAwaiting,
			Update: createdAt,
			In:     c.incoming,
			Id:     api.OfferId(offerId),
			Index:  -1,
		},
	}

	func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.mem.Save(offer)
		c.file.Save(offer)
	}()

	func() {
		c.Notify <- api.Notification{
			Offer:    *offer,
			Peer:     Peer(c.Core).Get(peerId),
			Incoming: c.incoming,
		}
		go c.notify(offer.Status, c.statusSubs)
		go c.notify(offer, c.offerSubs)
	}()
}

func (c Offer) Update(offer *api.Offer, index int) {
	offer.Index = index
	if index == len(offer.Files) {
		offer.Status.Progress = 0
	}
	ongoing := index > -1 && index < len(offer.Files)
	if ongoing {
		offer.Update = time.Now().UnixMilli()
	}

	func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.mem.Save(offer)
		if !ongoing {
			c.file.Save(offer)
		}
	}()

	func() {
		var info *api.Info
		if ongoing {
			info = &offer.Files[index]
		}
		c.Notify <- api.Notification{
			Incoming: c.incoming,
			Peer:     Peer(c.Core).Get(offer.Peer),
			Offer:    *offer,
			Info:     info,
		}
		go c.notify(offer.Status, c.statusSubs)
	}()
}

func (c Offer) Get(id api.OfferId) (offer *api.Offer) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	offer = c.mem.Get()[id]
	return
}

func (c Offer) List() (offers []api.Offer) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := c.mem.Get()
	for _, offer := range m {
		offers = append(offers, *offer)
	}
	sort.Sort(api.ByTimestamp(offers))
	return
}

func (c Offer) Status() *api.Subscriptions {
	return c.statusSubs
}

func (c Offer) Offers() *api.Subscriptions {
	return c.offerSubs
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
		subscriber <- jsonData
	}
}
