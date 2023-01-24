package optimizer

import (
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/node/peers"
)

func scoreAddr(addr infra.Addr) int {
	switch addr.Network() {
	case tor.NetworkName:
		return 10
	case bt.NetworkName:
		return 20
	case gw.NetworkName:
		return 30
	case inet.NetworkName:
		return 40
	}
	return 0
}

func scorePeer(peer *peers.Peer) (best int) {
	for _, link := range peer.Links() {
		score := scoreAddr(link.RemoteAddr())
		if score > best {
			best = score
		}
	}
	return best
}
