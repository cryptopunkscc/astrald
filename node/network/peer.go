package network

import (
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/astral/link/activity"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"io"
	"sync"
)

type Peer struct {
	*activity.Activity
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
		Activity: activity.New(nil),
	}
}

func (peer *Peer) Identity() id.Identity {
	return peer.id
}

func (peer *Peer) AddLink(link *link.Link) error {
	peer.linksMu.Lock()
	defer peer.linksMu.Unlock()

	err := peer.links.Add(link)
	if err != nil {
		return err
	}

	link.SetParent(peer)
	peer.Touch()

	if peer.links.Count() == 1 {
		peer.Activity.SetSticky(true)
		peer.triggerLinked()
	}

	go func() {
		<-link.WaitClose()
		_ = peer.removeLink(link)
	}()

	return nil
}

func (peer *Peer) PreferredLink() *link.Link {
	var best *link.Link

	for link := range peer.links.Each() {
		if best == nil {
			best = link
			continue
		}

		if best.Network() == tor.NetworkName {
			if link.Network() == inet.NetworkName {
				best = link
			}
		}
	}

	return best
}

func (peer *Peer) Query(query string) (io.ReadWriteCloser, error) {
	return peer.PreferredLink().Query(query)
}

func (peer *Peer) WaitLinked() <-chan struct{} {
	peer.linksMu.Lock()
	defer peer.linksMu.Unlock()

	return peer.linkedCh
}

func (peer *Peer) Links() <-chan *link.Link {
	return peer.links.Each()
}

func (peer *Peer) LinkCount() int {
	return peer.links.Count()
}

func (peer *Peer) removeLink(link *link.Link) error {
	peer.linksMu.Lock()
	defer peer.linksMu.Unlock()

	if err := peer.links.Remove(link); err != nil {
		return err
	}

	peer.AddBytesRead(link.BytesRead())
	peer.AddBytesWritten(link.BytesWritten())

	if peer.links.Count() == 0 {
		peer.resetLinked()
		peer.SetSticky(false)
	}
	return nil
}

func (peer *Peer) triggerLinked() {
	close(peer.linkedCh)
}

func (peer *Peer) resetLinked() {
	peer.linkedCh = make(chan struct{})
}
