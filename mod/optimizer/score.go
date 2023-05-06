package optimizer

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/network"
)

func scoreAddr(addr net.Endpoint) int {
	switch addr.Network() {
	case "tor":
		return 10
	case "bt":
		return 20
	case "gw":
		return 30
	case "inet":
		return 40
	}
	return 0
}

func scorePeer(peer *network.Peer) (best int) {
	for _, link := range peer.Links() {
		score := scoreAddr(link.RemoteEndpoint())
		if score > best {
			best = score
		}
	}
	return best
}
