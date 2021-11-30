package network

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
)

func (n *Network) handlePresence(presence infra.Presence) error {
	if presence.Identity.IsEqual(n.localID) {
		return nil
	}

	n.Contacts.AddAddr(presence.Identity, presence.Addr)

	// maintain links with present devices
	n.Connect(context.Background(), n.Peer(presence.Identity))
	return nil
}
