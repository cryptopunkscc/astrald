package peer

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"sync"
	"time"
)

var _ sig.Idler = &Peer{}

type Peer struct {
	id    id.Identity
	links map[*link.Link]struct{}
	queue *sig.Queue
	mu    sync.Mutex
}

func New(id id.Identity) *Peer {
	return &Peer{
		id:    id,
		links: make(map[*link.Link]struct{}),
		queue: &sig.Queue{},
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

	peer.queue = peer.queue.Push(link)

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

func (peer *Peer) FollowLinks(ctx context.Context, onlyNew bool) <-chan *link.Link {
	var ch chan *link.Link

	if onlyNew {
		ch = make(chan *link.Link)
	} else {
		peer.mu.Lock()
		ch = make(chan *link.Link, len(peer.links))
		for l := range peer.links {
			ch <- l
		}
		peer.mu.Unlock()
	}

	go func() {
		defer close(ch)
		for i := range peer.queue.Follow(ctx) {
			ch <- i.(*link.Link)
		}
	}()

	return ch
}

func (peer *Peer) Query(query string) (io.ReadWriteCloser, error) {
	queryLink := link.Select(peer.Links(), link.Fastest)

	if queryLink == nil {
		return nil, errors.New("no suitable link found")
	}

	return queryLink.Query(query)
}

func (peer *Peer) remove(link *link.Link) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if _, found := peer.links[link]; !found {
		return errors.New("not found")
	}

	delete(peer.links, link)

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
