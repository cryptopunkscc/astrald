package nat

import (
	"github.com/cryptopunkscc/astrald/net"
)

type inboundConn struct {
	net.Conn
}

func (inboundConn) Outbound() bool {
	return false
}
