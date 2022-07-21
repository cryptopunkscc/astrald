package service

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/file"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/memory"
	"log"
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

type OfferUpdates api.Core

func (core OfferUpdates) Start() func() {
	receive := make(chan api.OfferUpdate, 1024)
	core.Channel.Offers = receive

	go func() {
		buffer := newOfferUpdatesBuffer()
		for {
			select {
			case update := <-receive:
				// Add received update to buffer
				status := update.Stat()
				buffer[status.In][status.Id] = update
			default:
				switch {
				case buffer.len() == 0:
					// There are no elements to proceed.
					// Wait for next element and continue buffer update.
					receive <- <-receive
				default:
					log.Println("updates buffer size:", buffer.len())
					// Start processing buffered elements:
					// Prepare array with sorted updates.
					startTime := time.Now().UnixNano()
					var updates []api.OfferUpdate
					for _, b := range buffer {
						for _, next := range b {
							updates = append(updates, next)
						}
					}
					sort.Sort(api.OfferUpdatesByUpdate(updates))

					// Save updates in memory cache.
					core.Mutex.Incoming.Lock()
					core.Mutex.Outgoing.Lock()
					for _, update := range updates {
						update.Cache()
					}
					core.Mutex.Incoming.Unlock()
					core.Mutex.Outgoing.Unlock()

					// Save updates in storage.
					for _, update := range updates {
						update.Save()
					}

					// Notify listeners
					for _, update := range updates {
						update.Forward()
					}

					// Display system notification
					arr := make([]api.Notification, len(updates))
					for _, update := range updates {
						arr = append(arr, update.Notification())
					}
					core.Sys.Notify(arr)

					// Cleanup buffer
					buffer = newOfferUpdatesBuffer()
					endTime := time.Now().UnixNano()
					workTime := endTime - startTime
					sleepTime := int64(time.Second) - (workTime * int64(time.Nanosecond))
					time.Sleep(time.Nanosecond * time.Duration(sleepTime))
				}
			}
		}
	}()

	return func() {
		close(receive)
	}
}

type offerUpdatesBuffer map[bool]map[api.OfferId]api.OfferUpdate

func newOfferUpdatesBuffer() offerUpdatesBuffer {
	return offerUpdatesBuffer{true: {}, false: {}}
}

func (b offerUpdatesBuffer) len() int {
	return len(b[true]) + len(b[false])
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
	sort.Sort(api.OffersByCreate(offers))
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
		Create: time.Now().UnixNano(),
		OfferStatus: api.OfferStatus{
			Status: api.StatusAwaiting,
			In:     offer.incoming,
			Id:     api.OfferId(offerId),
			Index:  -1,
		},
	}
	offer.dispatch()
}

func (offer *Offer) Accept() {
	offer.Status = api.StatusAccepted
	offer.Index = -1
	offer.dispatch()
}

func (offer *Offer) Update(progress int64) {
	offer.Progress = progress
	offer.dispatch()
}

func (offer *Offer) Finish(err error) {
	if err == nil {
		offer.Index = len(offer.Files)
		offer.Progress = 0
		offer.Status = api.StatusCompleted
	} else {
		offer.Status = api.StatusFailed
	}
	offer.dispatch()
}

func (offer Offer) dispatch() {
	offer.Offer.Update = time.Now().UnixMicro()
	o := *offer.Offer
	offer.Offer = &o
	offer.Channel.Offers <- &offer
}

func (offer *Offer) Forward() {
	offer.notify(offer.OfferStatus, offer.StatusSubs)
	if offer.Status == api.StatusAwaiting {
		offer.notify(offer.Offer, offer.OfferSubs)
	}
}

func (offer *Offer) Notification() (n api.Notification) {
	n = api.Notification{
		Peer:  Peer(offer.Core).Get(offer.Peer),
		Offer: *offer.Offer,
	}
	if offer.IsOngoing() {
		n.Info = &offer.Files[offer.Index]
	}
	return
}

func (offer *Offer) Cache() {
	offer.mem.Save(*offer.Offer)
}

func (offer Offer) Save() {
	if !offer.IsOngoing() {
		offer.file.Save(*offer.Offer)
	}
}

func (offer *Offer) Stat() api.OfferStatus {
	return offer.OfferStatus
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
