package peer

import (
	"errors"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node/auth/id"
	_link "github.com/cryptopunkscc/astrald/node/link"
	"log"
	"sync"
)

type Peer struct {
	peerID   id.Identity
	links    []*_link.Link
	mu       sync.Mutex
	requests chan _link.Request
}

func New(peerID id.Identity) *Peer {
	return &Peer{
		peerID:   peerID,
		links:    make([]*_link.Link, 0),
		requests: make(chan _link.Request),
	}
}

func (peer *Peer) AddLink(link *_link.Link) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	log.Printf(
		"new %slink %s net %s addr %s\n",
		logfmt.Dir(link.Outbound()),
		logfmt.ID(link.RemoteIdentity().String()),
		link.RemoteAddr().Network(),
		link.RemoteAddr().String(),
	)

	// Safety check
	if link.RemoteIdentity().String() != peer.peerID.String() {
		return errors.New("identity mismatch")
	}

	// Check for duplicates
	for _, l := range peer.links {
		if link == l {
			return errors.New("already added")
		}
	}

	peer.links = append(peer.links, link)

	if len(peer.links) == 1 {
		log.Println("Peer", logfmt.ID(peer.peerID.String()), "online.")
	}

	go func() {
		for req := range link.Requests() {
			peer.requests <- req
		}
		err := peer.removeLink(link)
		if err != nil {
			log.Println("[Peer] error removing link from peer:", err)
		}
	}()

	return nil
}

// Requests returns a channel to which incoming requests will be sent
func (peer *Peer) Requests() <-chan _link.Request {
	return peer.requests
}

func (peer *Peer) Connected() bool {
	return len(peer.links) > 0
}

// DefaultLink returns the default link to be used for new connections
func (peer *Peer) DefaultLink() *_link.Link {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if len(peer.links) < 1 {
		return nil
	}

	return peer.links[0]
}

// removeLink removes a link from the link list
func (peer *Peer) removeLink(link *_link.Link) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	for i, l := range peer.links {
		if l == link {
			peer.links = append(peer.links[:i], peer.links[i+1:]...)

			log.Printf(
				"lost %slink %s %s\n",
				logfmt.Dir(link.Outbound()),
				logfmt.ID(link.RemoteIdentity().String()),
				link.RemoteAddr(),
			)

			if len(peer.links) == 0 {
				log.Println("Peer", logfmt.ID(peer.peerID.String()), "offline.")
			}
			return nil
		}
	}

	return errors.New("not found")
}
