package network

import (
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/logfmt"
	"log"
	"sync"
	"time"
)

type Peer struct {
	id           id.Identity
	links        *link.Set
	linksMu      sync.Mutex
	linkedCh     chan struct{}
	bytesRead    int
	bytesWritten int
}

func (peer *Peer) BytesRead() int {
	n := peer.bytesRead
	for link := range peer.Links() {
		n += link.BytesRead()
	}
	return n
}

func (peer *Peer) BytesWritten() int {
	n := peer.bytesWritten
	for link := range peer.Links() {
		n += link.BytesWritten()
	}
	return n
}

func (peer *Peer) ID() id.Identity {
	return peer.id
}

func NewPeer(id id.Identity) *Peer {
	return &Peer{
		id:       id,
		links:    link.NewSet(),
		linkedCh: make(chan struct{}),
	}
}

func (peer *Peer) AddLink(link *link.Link) error {
	peer.linksMu.Lock()
	defer peer.linksMu.Unlock()

	err := peer.links.Add(link)
	if err != nil {
		return err
	}

	if peer.links.Count() == 1 {
		log.Println(logfmt.ID(peer.id), "linked")
		peer.triggerLinked()
	}

	go func() {
		<-link.WaitClose()
		_ = peer.removeLink(link)
	}()

	return nil
}

func (peer *Peer) PreferredLink() *link.Link {
	if peer.links.Count() == 0 {
		return nil
	}

	return <-peer.links.Each()
}

func (peer *Peer) Idle() time.Duration {
	if peer.links.Count() == 0 {
		return -1
	}

	links := peer.links.Each()
	best := (<-links).Idle()
	for link := range links {
		idle := link.Idle()
		if idle < best {
			best = idle
		}
	}

	return best
}

func (peer *Peer) WaitLinked() <-chan struct{} {
	peer.linksMu.Lock()
	defer peer.linksMu.Unlock()

	return peer.linkedCh
}

func (peer *Peer) Links() <-chan *link.Link {
	return peer.links.Each()
}

func (peer *Peer) removeLink(link *link.Link) error {
	peer.linksMu.Lock()
	defer peer.linksMu.Unlock()

	if err := peer.links.Remove(link); err != nil {
		return err
	}

	peer.bytesRead += link.BytesRead()
	peer.bytesWritten += link.BytesWritten()

	if peer.links.Count() == 0 {
		log.Println(logfmt.ID(peer.id), "unlinked")
		peer.resetLinked()
	}
	return nil
}

func (peer *Peer) triggerLinked() {
	close(peer.linkedCh)
}

func (peer *Peer) resetLinked() {
	peer.linkedCh = make(chan struct{})
}
