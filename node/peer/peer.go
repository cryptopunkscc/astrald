package peer

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
	"time"
)

var _ sig.Idler = &Peer{}

type Peer struct {
	id     id.Identity
	links  map[*link.Link]struct{}
	events event.Queue
	mu     sync.Mutex
}

func New(id id.Identity) *Peer {
	return &Peer{
		id:    id,
		links: make(map[*link.Link]struct{}),
	}
}

func (peer *Peer) Identity() id.Identity {
	return peer.id
}

func (peer *Peer) Add(link *link.Link) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if _, found := peer.links[link]; found {
		return errors.New("duplicate link")
	}

	peer.links[link] = struct{}{}
	link.SetEventParent(&peer.events)

	if len(peer.links) == 1 {
		peer.events.Emit(EventLinked{peer, link})
	}

	peer.events.Emit(EventLinkEstablished{Link: link})

	go func() {
		<-link.Wait()
		peer.remove(link)
	}()

	return nil
}

func (peer *Peer) Links() <-chan *link.Link {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return peer.getLinks()
}

func (peer *Peer) Idle() time.Duration {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	l := link.Select(peer.getLinks(), link.MostRecent)
	if l == nil {
		return -1
	}
	return l.Idle()
}

func (peer *Peer) Subscribe(cancel sig.Signal) <-chan event.Event {
	return peer.events.Subscribe(cancel)
}

func (peer *Peer) SubscribeLinks(cancel sig.Signal, fetchAll bool) <-chan *link.Link {
	var ch chan *link.Link

	if fetchAll {
		peer.mu.Lock()
		ch = make(chan *link.Link, len(peer.links))
		for l := range peer.links {
			ch <- l
		}
		peer.mu.Unlock()
	} else {
		ch = make(chan *link.Link)
	}

	go func() {
		defer close(ch)
		for event := range peer.events.Subscribe(cancel) {
			if event, ok := event.(EventLinkEstablished); ok {
				ch <- event.Link
			}
		}
	}()

	return ch
}

func (peer *Peer) Query(ctx context.Context, query string) (*link.Conn, error) {
	queryLink := link.Select(peer.Links(), link.LowestRoundTrip)

	if queryLink == nil {
		return nil, errors.New("no link found")
	}

	return queryLink.Query(ctx, query)
}

func (peer *Peer) Unlink() {
	for l := range peer.Links() {
		l.Close()
	}
}

func (peer *Peer) remove(link *link.Link) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if _, found := peer.links[link]; !found {
		return errors.New("not found")
	}

	delete(peer.links, link)

	peer.events.Emit(EventLinkClosed{Link: link})
	if len(peer.links) == 0 {
		peer.events.Emit(EventUnlinked{peer})
	}

	return nil
}

func (peer *Peer) getLinks() <-chan *link.Link {
	ch := make(chan *link.Link, len(peer.links))
	for link, _ := range peer.links {
		ch <- link
	}
	close(ch)
	return ch
}

func (peer *Peer) WaitLinked(ctx context.Context) (*link.Link, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	select {
	case l, ok := <-peer.SubscribeLinks(ctx.Done(), true):
		if !ok {
			return nil, context.Canceled
		}
		return l, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
