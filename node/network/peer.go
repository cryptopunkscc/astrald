package network

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
	"sync/atomic"
	"time"
)

type Peer struct {
	id       id.Identity
	links    *LinkSet
	events   event.Queue
	done     chan struct{}
	unlinked atomic.Bool
	mu       sync.Mutex
}

// newPeer returns a new Peer instance.
func newPeer(id id.Identity, eventParent *event.Queue) *Peer {
	p := &Peer{
		id:    id,
		links: NewLinkSet(),
		done:  make(chan struct{}),
	}

	p.events.SetParent(eventParent)

	return p
}

// Identity returns the identity of the peer.
func (peer *Peer) Identity() id.Identity {
	return peer.id
}

// Links retruns an array of peer's links.
func (peer *Peer) Links() []*link.Link {
	return peer.links.All()
}

// PreferredLink returns the preferred link
func (peer *Peer) PreferredLink() *link.Link {
	return link.Select(peer.Links(), BestQuality)
}

// Check checks health of every link with the peer.
func (peer *Peer) Check() {
	for _, l := range peer.Links() {
		l.Health().Check()
	}
}

// Unlink closes all links with the peer.
func (peer *Peer) Unlink() {
	for _, link := range peer.links.All() {
		link.Close()
	}
}

// Unlinked returns true if the peer has been unlinked.
func (peer *Peer) Unlinked() bool {
	return peer.unlinked.Load()
}

// Idle returns the idle time of the most recently active link. Returns -1 if the peer has no links.
func (peer *Peer) Idle() time.Duration {
	l := link.Select(peer.Links(), link.MostRecent)
	if l == nil {
		return -1
	}
	return l.Activity().Idle()
}

// Done returns a channel that will be closed when the peer becomes unlinked.
func (peer *Peer) Done() <-chan struct{} {
	return peer.done
}

// Events returns the event queue of the peer
func (peer *Peer) Events() *event.Queue {
	return &peer.events
}

// setUnlinked marks the peer as unlinked. Returns false if the peer was already unlinked, true otherwise.
func (peer *Peer) setUnlinked() bool {
	if peer.unlinked.CompareAndSwap(false, true) {
		close(peer.done)
		return true
	}
	return false
}

// addLink adds a link to the peer. Peer cannot be unlinked.
func (peer *Peer) addLink(l *link.Link) (err error) {
	if peer.unlinked.Load() {
		return ErrPeerUnlinked
	}

	if err = peer.links.Add(l); err == nil {
		l.Events().SetParent(&peer.events)
	}
	return
}

func (peer *Peer) removeLink(l *link.Link) (err error) {
	if err = peer.links.Remove(l); err == nil {
		l.Events().SetParent(nil)
	}
	return
}
