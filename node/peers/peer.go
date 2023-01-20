package peers

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
	"time"
)

const MaxPeerLinks = 8

type Peer struct {
	id           id.Identity
	links        map[*link.Link]struct{}
	queries      chan *link.Query
	mu           sync.Mutex
	events       event.Queue
	done         chan struct{}
	disconnected bool
}

func NewPeer(id id.Identity, eventParent *event.Queue) *Peer {
	p := &Peer{
		id:      id,
		links:   make(map[*link.Link]struct{}),
		queries: make(chan *link.Query),
		done:    make(chan struct{}),
	}

	p.events.SetParent(eventParent)

	return p
}

func (peer *Peer) Identity() id.Identity {
	return peer.id
}

func (peer *Peer) Queries() <-chan *link.Query {
	return peer.queries
}

func (peer *Peer) Idle() time.Duration {
	l := link.Select(peer.Links(), link.MostRecent)
	if l == nil {
		return -1
	}
	return l.Idle()
}

func (peer *Peer) Links() <-chan *link.Link {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	ch := make(chan *link.Link, len(peer.links))
	for l := range peer.links {
		ch <- l
	}
	close(ch)
	return ch
}

func (peer *Peer) LinkCount() int {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return len(peer.links)
}

func (peer *Peer) PreferredLink() *link.Link {
	return link.Select(peer.Links(), link.LowestRoundTrip)
}

func (peer *Peer) Unlink() {
	for link := range peer.Links() {
		link.Close()
	}
}

func (peer *Peer) Wait() <-chan struct{} {
	return peer.done
}

func (peer *Peer) addLink(l *link.Link) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.disconnected {
		return errors.New("peer disconnected")
	}

	if len(peer.links) >= MaxPeerLinks {
		return errors.New("link limit exceeded")
	}

	if _, found := peer.links[l]; found {
		return errors.New("duplicate link")
	}

	l.SetEventParent(&peer.events)
	peer.links[l] = struct{}{}

	go func() {
		for q := range l.Queries() {
			peer.queries <- q
		}
	}()

	return nil
}

func (peer *Peer) removeLink(l *link.Link) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if _, found := peer.links[l]; !found {
		return errors.New("not found")
	}

	delete(peer.links, l)

	if len(peer.links) == 0 {
		close(peer.queries)
		close(peer.done)
		peer.disconnected = true
	}

	return nil
}
