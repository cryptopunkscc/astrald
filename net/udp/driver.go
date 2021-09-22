package udp

import (
	"github.com/cryptopunkscc/astrald/net"
)

type driver struct{}

var _ net.BroadcastNetwork = &driver{}

func init() {
	if err := net.AddBroadcastNetwork("udp", &driver{}); err != nil {
		panic(err)
	}
}
