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
	links map[*link.Link]struct{}
	id    id.Identity
	mu    sync.Mutex
	queue *sig.Queue
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

	peer.queue.Push(nil)

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

func (peer *Peer) getLinks() <-chan *link.Link {
	ch := make(chan *link.Link, len(peer.links))
	for link, _ := range peer.links {
		ch <- link
	}
	close(ch)
	return ch
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

func (peer *Peer) StateQueue() *sig.Queue {
	return peer.queue
}

func (peer *Peer) Follow(ctx context.Context) <-chan interface{} {
	return peer.queue.Follow(ctx)
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

	peer.queue.Push(nil)

	return nil
}
