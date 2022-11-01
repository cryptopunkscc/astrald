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
		mu:         &c.Mutex.Outgoing,
		mem:        memory.Outgoing(c),
		file:       file.Outgoing(c),
		offerSubs:  c.Observers.OutgoingOffers,
		statusSubs: c.Observers.OutgoingStatus,
	}
}

func Incoming(c core.Component) *Offer {
	return &Offer{
		Component:  c,
		mu:         &c.Mutex.Incoming,
		mem:        memory.Incoming(c),
		file:       file.Incoming(c),
		offerSubs:  c.Observers.IncomingOffers,
		statusSubs: c.Observers.IncomingStatus,
		incoming:   true,
	}
}

type Offer struct {
	core.Component
	*warpdrive.Offer
	mu         *sync.RWMutex
	offerSubs  *warpdrive.Subscriptions
	statusSubs *warpdrive.Subscriptions
	file       storage.Offer
	mem        storage.Offer
	incoming   bool
}

var _ warpdrive.OfferService = &Offer{}

func (srv *Offer) OfferSubscriptions() *warpdrive.Subscriptions {
	return srv.offerSubs
}

func (srv *Offer) StatusSubscriptions() *warpdrive.Subscriptions {
	return srv.statusSubs
}

func (srv *Offer) Get(id warpdrive.OfferId) (offer *warpdrive.Offer) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
	offer = srv.mem.Get()[id]
	return
}

func (srv *Offer) List() (offers []warpdrive.Offer) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
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

func (srv *Offer) update(offer *warpdrive.Offer, progress int64) {
	offer.Progress = progress
	srv.dispatch(offer)
}

func (srv *Offer) forward() {
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

func (srv *Offer) notification() (n notify.Notification) {
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
	receive := make(chan interface{}, 1024)
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
			case received := <-receive:
				update := received.(*Offer)
				if update == nil {
					return
				}
				// Add received update to buffer
				status := update.OfferStatus
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
	var updates []*Offer
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
		update.mem.Save(*update.Offer)
	}
	c.Mutex.Incoming.Unlock()
	c.Mutex.Outgoing.Unlock()

	// Save updates in storage.
	for _, update := range updates {
		if !update.IsOngoing() {
			update.file.Save(*update.Offer)
		}
	}

	// Notify listeners
	for _, update := range updates {
		update.forward()
	}

	// Display system notification
	arr := make([]notify.Notification, len(updates))
	for i, update := range updates {
		arr[i] = update.notification()
	}
	c.Sys.Notify(arr)
}

type offerUpdatesBuffer map[bool]map[warpdrive.OfferId]*Offer

func newOfferUpdatesBuffer() offerUpdatesBuffer {
	return offerUpdatesBuffer{true: {}, false: {}}
}

func (b offerUpdatesBuffer) len() int {
	return len(b[true]) + len(b[false])
}

type byUpdateTime []*Offer

func (a byUpdateTime) Len() int           { return len(a) }
func (a byUpdateTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byUpdateTime) Less(i, j int) bool { return a[i].Update < a[j].Update }
