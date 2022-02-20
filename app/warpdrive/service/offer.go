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
	*api.Offer
	*sync.RWMutex
	OfferSubs  *api.Subscriptions
	StatusSubs *api.Subscriptions
	file       api.OfferStorage
	mem        api.OfferStorage
	incoming   bool
}

func Outgoing(c api.Core) *Offer {
	return &Offer{
		Core:       c,
		RWMutex:    &c.Mutex.Outgoing,
		mem:        memory.Outgoing(c),
		file:       file.Outgoing(c),
		OfferSubs:  c.Observers.OutgoingOffers,
		StatusSubs: c.Observers.OutgoingStatus,
	}
}

func Incoming(c api.Core) *Offer {
	return &Offer{
		Core:       c,
		RWMutex:    &c.Mutex.Incoming,
		mem:        memory.Incoming(c),
		file:       file.Incoming(c),
		OfferSubs:  c.Observers.IncomingOffers,
		StatusSubs: c.Observers.IncomingStatus,
		incoming:   true,
	}
}

func (offer *Offer) List() (offers []api.Offer) {
	offer.RLock()
	defer offer.RUnlock()
	m := offer.mem.Get()
	for _, o := range m {
		offers = append(offers, *o)
	}
	sort.Sort(api.ByTimestamp(offers))
	return
}

func (offer *Offer) Set(id api.OfferId) *Offer {
	offer.RLock()
	defer offer.RUnlock()
	offer.Offer = offer.mem.Get()[id]
	return offer
}

func (offer *Offer) Add(offerId string, files []api.Info, peerId api.PeerId) {
	offer.Offer = &api.Offer{
		Files:  files,
		Peer:   peerId,
		Create: time.Now().UnixMilli(),
		OfferStatus: api.OfferStatus{
			Status: api.StatusAwaiting,
			In:     offer.incoming,
			Id:     api.OfferId(offerId),
			Index:  -1,
		},
	}
	offer.commit()
}

func (offer *Offer) Accept() {
	offer.Status = api.StatusAccepted
	offer.Index = -1
	offer.commit()
}

func (offer *Offer) Update(progress int64) {
	offer.Progress = progress
	offer.commit()
}

func (offer *Offer) Finish(err error) {
	if err == nil {
		offer.Index = len(offer.Files)
		offer.Progress = 0
		offer.Status = api.StatusCompleted
	} else {
		offer.Status = api.StatusFailed
	}
	offer.commit()
}

func (offer *Offer) commit() {
	offer.Offer.Update = time.Now().UnixMilli()

	ongoing := offer.Index > -1 && offer.Index < len(offer.Files)

	// save
	func() {
		offer.Lock()
		defer offer.Unlock()
		offer.mem.Save(offer.Offer)
		if ongoing {
			offer.file.Save(offer.Offer)
		}
	}()

	// notify
	func() {
		var info *api.Info
		if ongoing {
			info = &offer.Files[offer.Index]
		}
		offer.Notify <- api.Notification{
			Incoming: offer.incoming,
			Peer:     Peer(offer.Core).Get(offer.Peer),
			Offer:    *offer.Offer,
			Info:     info,
		}
		go offer.notify(offer.OfferStatus, offer.StatusSubs)
		if offer.Status == api.StatusAwaiting {
			go offer.notify(offer.Offer, offer.OfferSubs)
		}
	}()
}

func (offer *Offer) notify(data interface{}, subscribers *api.Subscriptions) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		offer.Println("Cannot create json from data", data, err)
		return
	}
	subscribers.Lock()
	defer subscribers.Unlock()
	for subscriber := range subscribers.Set {
		subscriber <- jsonData
	}
}
