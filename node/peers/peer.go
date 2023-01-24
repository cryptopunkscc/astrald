package peers

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
	"time"
)

type Peer struct {
	Events event.Queue

	id       id.Identity
	links    map[*link.Link]struct{}
	queries  chan *link.Query
	done     chan struct{}
	unlinked bool
	mu       sync.Mutex
}

// newPeer returns a new Peer instance.
func newPeer(id id.Identity, eventParent *event.Queue) *Peer {
	p := &Peer{
		id:      id,
		links:   make(map[*link.Link]struct{}),
		queries: make(chan *link.Query),
		done:    make(chan struct{}),
	}

	p.Events.SetParent(eventParent)

	return p
}

// Identity returns the identity of the peer.
func (peer *Peer) Identity() id.Identity {
	return peer.id
}

// Queries returns a channel receiving peer queries. The channel will be closed when the peer gets unlinked.
func (peer *Peer) Queries() <-chan *link.Query {
	return peer.queries
}

// Links retruns an array of peer's links.
func (peer *Peer) Links() []*link.Link {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	links := make([]*link.Link, 0, len(peer.links))
	for l := range peer.links {
		links = append(links, l)
	}

	return links
}

// PreferredLink returns the preferred link
func (peer *Peer) PreferredLink() *link.Link {
	return link.Select(peer.Links(), link.BestQuality)
}

// Unlink closes all links with the peer. After a peer is unlinked, no links can be added to it.
func (peer *Peer) Unlink() {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	peer.unlink()

	for link := range peer.links {
		link.Close()
	}
}

// Unlinked returns true if the peer has been unlinked.
func (peer *Peer) Unlinked() bool {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return peer.unlinked
}

// Idle returns the idle time of the most recently active link. Returns -1 if the peer has no links.
func (peer *Peer) Idle() time.Duration {
	l := link.Select(peer.Links(), link.MostRecent)
	if l == nil {
		return -1
	}
	return l.Idle()
}

// Done returns a channel that will be closed when the peer becomes unlinked.
func (peer *Peer) Done() <-chan struct{} {
	return peer.done
}

// unlink marks the peer as unlinked and closes its output channels.
func (peer *Peer) unlink() {
	if peer.unlinked {
		return
	}
	peer.unlinked = true
	close(peer.queries)
	close(peer.done)
}

// addLink adds a link to the peer
// Errors: ErrPeerUnlinked, ErrPeerLinkLimitExceeded, ErrDuplicateLink
func (peer *Peer) addLink(l *link.Link) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.unlinked {
		return ErrPeerUnlinked
	}

	if len(peer.links) >= MaxPeerLinks {
		return ErrPeerLinkLimitExceeded
	}

	if _, found := peer.links[l]; found {
		return ErrDuplicateLink
	}

	l.SetEventParent(&peer.Events)
	peer.links[l] = struct{}{}

	go func() {
		for q := range l.Queries() {
			peer.queries <- q
		}
	}()

	return nil
}

// removeLink removes a link from the peer
// Errors: ErrLinkNotFound
func (peer *Peer) removeLink(l *link.Link) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if _, found := peer.links[l]; !found {
		return ErrLinkNotFound
	}

	delete(peer.links, l)

	if len(peer.links) == 0 {
		peer.unlink()
	}

	return nil
}
