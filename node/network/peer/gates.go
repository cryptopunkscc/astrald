package peer

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/sig"
)

// LinkedGate is open whenever the peer is linked
func LinkedGate(ctx context.Context, peer *Peer) *sig.Gate {
	gate := &sig.Gate{}

	go func() {
		for {
			if len(peer.Links.Links()) > 0 {
				gate.Open()
			} else {
				gate.Close()
			}

			select {
			case <-peer.Links.Wait():
			case <-ctx.Done():
				return
			}
		}
	}()

	return gate
}

// NetworkUnlinkedGate is open whenever the peer has no links over the specified network
func NetworkUnlinkedGate(ctx context.Context, peer *Peer, network string) *sig.Gate {
	gate := &sig.Gate{}

	go func() {
		q := peer.Links.Queue
		for {
			if len(link.Filter(peer.Links.Links(), link.Network(network))) == 0 {
				gate.Open()
			} else {
				gate.Close()
			}

			select {
			case <-q.Wait():
				if q = q.Next(); q == nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return gate
}
