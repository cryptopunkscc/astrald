package node

import (
	"context"
	"errors"
	alink "github.com/cryptopunkscc/astrald/link"
	nlink "github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/sig"
)

func (node *Node) AddLink(link *alink.Link) error {
	if link == nil {
		return errors.New("link is nil")
	}

	node.links <- link

	return nil
}

func (node *Node) processLinks(ctx context.Context) {
	for {
		select {
		case link := <-node.links:
			node.addLink(ctx, nlink.Wrap(link))
		case <-ctx.Done():
			return
		}
	}
}

func (node *Node) addLink(ctx context.Context, link *nlink.Link) error {
	peer := node.Peers.Find(link.RemoteIdentity(), true)

	if err := peer.Add(link); err != nil {
		return err
	}

	// forward link's requests
	go func() {
		for query := range link.Queries() {
			node.queries <- query
		}
		node.emitEvent(EventLinkDown{link})

		if len(peer.Links()) == 0 {
			node.emitEvent(EventPeerUnlinked{peer})
		}
	}()

	node.emitEvent(EventLinkUp{link})

	if len(peer.Links()) == 1 {
		node.emitEvent(EventPeerLinked{peer, link})

		// set a timeout
		sig.On(ctx, sig.Idle(ctx, peer, defaultPeerIdleTimeout), func() {
			for l := range peer.Links() {
				l.Close()
			}
		})
	}

	return nil
}
