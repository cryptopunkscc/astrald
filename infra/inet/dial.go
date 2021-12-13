package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"net"
)

func Dial(ctx context.Context, addr Addr) (infra.Conn, error) {
	var d net.Dialer

	tcpConn, err := d.DialContext(ctx, "tcp", addr.String())
	if err != nil {
		return nil, err
	}

	return newConn(tcpConn, true), nil
}
