package peer

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"sync"
	"time"
)

var _ sig.Idler = &Peer{}

type Peer struct {
	Links *link.Set
	id    id.Identity
	mu    sync.Mutex
}

func New(id id.Identity) *Peer {
	return &Peer{
		id:    id,
		Links: link.NewSet(),
	}
}

func (peer *Peer) Identity() id.Identity {
	return peer.id
}

func (peer *Peer) Add(link *link.Link) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	// add link to the set
	if err := peer.Links.Add(link); err != nil {
		return err
	}

	sig.On(context.Background(), link, func() {
		_ = peer.removeLink(link)
	})

	return nil
}

func (peer *Peer) Idle() time.Duration {
	if l := link.Select(peer.Links.Links(), link.MostRecent); l != nil {
		return l.Idle()
	}

	return 0
}

func (peer *Peer) Query(query string) (io.ReadWriteCloser, error) {
	queryLink := link.Select(peer.Links.Links(), link.Fastest)

	if queryLink == nil {
		return nil, errors.New("no suitable link found")
	}

	return queryLink.Query(query)
}

func (peer *Peer) removeLink(link *link.Link) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	// remove link from the set
	if err := peer.Links.Remove(link); err != nil {
		return err
	}

	return nil
}
