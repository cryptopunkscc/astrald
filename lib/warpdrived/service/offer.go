package service

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/core"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/notify"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/storage"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/storage/file"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/storage/memory"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"sort"
	"sync"
	"time"
)

func Outgoing(c core.Component) *Offer {
	return &Offer{
		Component:  c,
		RWMutex:    &c.Mutex.Outgoing,
		mem:        memory.Outgoing(c),
		file:       file.Outgoing(c),
		OfferSubs:  c.Observers.OutgoingOffers,
		StatusSubs: c.Observers.OutgoingStatus,
	}
}

func Incoming(c core.Component) *Offer {
	return &Offer{
		Component:  c,
		RWMutex:    &c.Mutex.Incoming,
		mem:        memory.Incoming(c),
		file:       file.Incoming(c),
		OfferSubs:  c.Observers.IncomingOffers,
		StatusSubs: c.Observers.IncomingStatus,
		incoming:   true,
	}
}

type Offer struct {
	core.Component
	*warpdrive.Offer
	*sync.RWMutex
	OfferSubs  *warpdrive.Subscriptions
	StatusSubs *warpdrive.Subscriptions
	file       storage.Offer
	mem        storage.Offer
	incoming   bool
}

var _ warpdrive.OfferService = &Offer{}

func (srv *Offer) List() (offers []warpdrive.Offer) {
	srv.RLock()
	defer srv.RUnlock()
	m := srv.mem.Get()
	for _, o := range m {
		offers = append(offers, *o)
	}
	sort.Sort(byCreate(offers))
	return
}

type byCreate []warpdrive.Offer

func (a byCreate) Len() int           { return len(a) }
func (a byCreate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byCreate) Less(i, j int) bool { return a[i].Create < a[j].Create }

func (srv *Offer) Get(id warpdrive.OfferId) (offer *warpdrive.Offer) {
	srv.RLock()
	defer srv.RUnlock()
	offer = srv.mem.Get()[id]
	return
}

func (srv *Offer) Add(
	offerId warpdrive.OfferId,
	files []warpdrive.Info,
	peerId warpdrive.PeerId,
) (offer *warpdrive.Offer) {
	offer = &warpdrive.Offer{
		Files:  files,
		Peer:   peerId,
		Create: time.Now().UnixMilli(),
		OfferStatus: warpdrive.OfferStatus{
			Status: warpdrive.StatusAwaiting,
			In:     srv.incoming,
			Id:     offerId,
			Index:  -1,
		},
	}
	srv.dispatch(offer)
	return
}

func (srv *Offer) Accept(offer *warpdrive.Offer) {
	offer.Status = warpdrive.StatusAccepted
	srv.dispatch(offer)
}

func (srv *Offer) Update(offer *warpdrive.Offer, progress int64) {
	offer.Progress = progress
	srv.dispatch(offer)
}

func (srv *Offer) Finish(offer *warpdrive.Offer, err error) {
	if err == nil {
		offer.Index = len(offer.Files)
		offer.Progress = 0
		offer.Status = warpdrive.StatusCompleted
	} else {
		offer.Status = warpdrive.StatusFailed
	}
	srv.dispatch(offer)
}

func (srv *Offer) dispatch(offer *warpdrive.Offer) {
	offer.Update = time.Now().UnixMilli()
	srv.Offer = offer
	srv.Channel.Offers <- srv
}

func (srv *Offer) OfferSubscriptions() *warpdrive.Subscriptions {
	return srv.OfferSubs
}

func (srv *Offer) StatusSubscriptions() *warpdrive.Subscriptions {
	return srv.StatusSubs
}

var _ core.OfferUpdate = &Offer{}

func (srv *Offer) Cache() {
	srv.mem.Save(*srv.Offer)
}

func (srv *Offer) Save() {
	if !srv.IsOngoing() {
		srv.file.Save(*srv.Offer)
	}
}

func (srv *Offer) Forward() {
	srv.notify(srv.OfferStatus, srv.StatusSubscriptions())
	if srv.Status == warpdrive.StatusAwaiting {
		srv.notify(srv.Offer, srv.OfferSubscriptions())
	}
}

func (srv *Offer) notify(data interface{}, subscribers *warpdrive.Subscriptions) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		srv.Println("Cannot create json from data", data, err)
		return
	}
	subscribers.Notify(jsonData)
}

func (srv *Offer) Stat() warpdrive.OfferStatus {
	return srv.OfferStatus
}

func (srv *Offer) Notification() (n notify.Notification) {
	n = notify.Notification{
		Peer:  Peer(srv.Component).Get(srv.Peer),
		Offer: *srv.Offer,
	}
	if srv.IsOngoing() {
		n.Info = &srv.Files[srv.Index]
	}
	return
}

type OfferUpdates core.Component

func (c OfferUpdates) Start(ctx context.Context) <-chan struct{} {
	receive := make(chan core.OfferUpdate, 1024)
	c.Channel.Offers = receive
	finish := make(chan struct{})
	go func() {
		buffer := newOfferUpdatesBuffer()
		defer func() {
			if buffer.len() > 0 {
				c.processUpdates(buffer)
			}
			close(finish)
		}()
		for {
			select {
			case update := <-receive:
				if update == nil {
					return
				}
				// Add received update to buffer
				status := update.Stat()
				buffer[status.In][status.Id] = update
			default:
				switch {
				case buffer.len() == 0:
					// There are no elements to proceed.
					// Wait for next element and continue buffer update.
					r := <-receive
					if r == nil {
						return
					}
					receive <- r
				default:
					// Start processing buffered elements:
					// Prepare array with sorted updates.
					startTime := time.Now().UnixNano()

					c.processUpdates(buffer)

					// Cleanup buffer
					buffer = newOfferUpdatesBuffer()
					// Measure work/sleep cycle to 1 second
					endTime := time.Now().UnixNano()
					workTime := endTime - startTime
					sleepTime := int64(time.Second) - workTime
					time.Sleep(time.Nanosecond * time.Duration(sleepTime))
				}
			}
		}
	}()

	go func() {
		<-ctx.Done()
		c.Job.Wait()
		time.Sleep(200 * time.Millisecond)
		close(receive)
		<-finish
		c.Println("finish updating offers")
	}()
	return finish
}

func (c OfferUpdates) processUpdates(buffer offerUpdatesBuffer) {
	var updates []core.OfferUpdate
	for _, b := range buffer {
		for _, next := range b {
			updates = append(updates, next)
		}
	}

	sort.Sort(byUpdateTime(updates))

	// Save updates in memory cache.
	c.Mutex.Incoming.Lock()
	c.Mutex.Outgoing.Lock()
	for _, update := range updates {
		update.Cache()
	}
	c.Mutex.Incoming.Unlock()
	c.Mutex.Outgoing.Unlock()

	// Save updates in storage.
	for _, update := range updates {
		update.Save()
	}

	// Notify listeners
	for _, update := range updates {
		update.Forward()
	}

	// Display system notification
	arr := make([]notify.Notification, len(updates))
	for i, update := range updates {
		arr[i] = update.Notification()
	}
	c.Sys.Notify(arr)
}

type offerUpdatesBuffer map[bool]map[warpdrive.OfferId]core.OfferUpdate

func newOfferUpdatesBuffer() offerUpdatesBuffer {
	return offerUpdatesBuffer{true: {}, false: {}}
}

func (b offerUpdatesBuffer) len() int {
	return len(b[true]) + len(b[false])
}

type byUpdateTime []core.OfferUpdate

func (a byUpdateTime) Len() int           { return len(a) }
func (a byUpdateTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byUpdateTime) Less(i, j int) bool { return a[i].Stat().Update < a[j].Stat().Update }
