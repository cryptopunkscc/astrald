package network

import (
	"github.com/cryptopunkscc/astrald/infra"
)

func (network *Network) handlePresence(presence infra.Presence) error {
	if presence.Identity.IsEqual(network.localID) {
		return nil
	}

	network.Graph.AddAddr(presence.Identity, presence.Addr)

	// maintain links with present devices
	network.Linker.Wake(presence.Identity)
	return nil
}
