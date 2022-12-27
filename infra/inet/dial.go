package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"net"
)

func (inet *Inet) Dial(ctx context.Context, addr Addr) (conn infra.Conn, err error) {
	var dialer net.Dialer
	var tcpConn net.Conn

	tcpConn, err = dialer.DialContext(ctx, "tcp", addr.String())
	if err != nil {
		return
	}

	conn = newConn(tcpConn, true)

	return
}
