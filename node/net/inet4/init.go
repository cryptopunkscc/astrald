package inet4

import (
	"github.com/cryptopunkscc/astrald/node/net"
)

func init() {
	net.Register("inet4", Dial)
}
