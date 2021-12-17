package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"net"
)

func (inet Inet) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	if _, ok := addr.(Addr); !ok {
		return nil, infra.ErrUnsupportedAddress
	}

	var dialer net.Dialer

	tcpConn, err := dialer.DialContext(ctx, "tcp", addr.String())
	if err != nil {
		return nil, err
	}

	return newConn(tcpConn, true), nil
}
