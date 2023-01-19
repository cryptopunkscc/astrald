package nat

import "github.com/cryptopunkscc/astrald/infra"

type inboundConn struct {
	infra.Conn
}

func (inboundConn) Outbound() bool {
	return false
}
