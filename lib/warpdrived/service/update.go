package service

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/core"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/notify"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"sort"
	"time"
)

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
				update := received.(*offer)
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
	var updates []*offer
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

type offerUpdatesBuffer map[bool]map[warpdrive.OfferId]*offer

func newOfferUpdatesBuffer() offerUpdatesBuffer {
	return offerUpdatesBuffer{true: {}, false: {}}
}

func (b offerUpdatesBuffer) len() int {
	return len(b[true]) + len(b[false])
}

type byUpdateTime []*offer

func (a byUpdateTime) Len() int           { return len(a) }
func (a byUpdateTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byUpdateTime) Less(i, j int) bool { return a[i].Update < a[j].Update }

func (srv *offer) forward() {
	srv.notify(srv.OfferStatus, srv.StatusSubscriptions())
	if srv.Status == warpdrive.StatusAwaiting {
		srv.notify(srv.Offer, srv.OfferSubscriptions())
	}
}

func (srv *offer) notify(data interface{}, subscribers *warpdrive.Subscriptions) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		srv.Println("Cannot create json from data", data, err)
		return
	}
	subscribers.Notify(jsonData)
}

func (srv *offer) notification() (n notify.Notification) {
	n = notify.Notification{
		Peer:  peer(srv.Component).Get(srv.Peer),
		Offer: *srv.Offer,
	}
	if srv.IsOngoing() {
		n.Info = &srv.Files[srv.Index]
	}
	return
}
