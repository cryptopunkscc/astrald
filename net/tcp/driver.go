package tcp

import (
	"github.com/cryptopunkscc/astrald/net"
)

type driver struct{}

var _ net.UnicastNetwork = &driver{}

func init() {
	if err := net.AddUnicastNetwork("tcp", &driver{}); err != nil {
		panic(err)
	}
}
