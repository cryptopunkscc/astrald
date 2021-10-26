package node

import (
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/logfmt"
	"log"
	"sync"
)

type Peer struct {
	id       id.Identity
	links    *link.Set
	linksMu  sync.Mutex
	linkedCh chan struct{}
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

func (peer *Peer) removeLink(link *link.Link) error {
	peer.linksMu.Lock()
	defer peer.linksMu.Unlock()

	if err := peer.links.Remove(link); err != nil {
		return err
	}

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

func (peer *Peer) WaitLinked() <-chan struct{} {
	peer.linksMu.Lock()
	defer peer.linksMu.Unlock()

	return peer.linkedCh
}
