package tor

import (
	"github.com/cryptopunkscc/astrald/node/net"
)

type driver struct{}

var _ net.UnicastNetwork = &driver{}

func init() {
	if err := net.AddUnicastNetwork("tor", &driver{}); err != nil {
		panic(err)
	}
}
