package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"golang.org/x/net/proxy"
)

func (drv *driver) Dial(_ context.Context, addr net.Addr) (net.Conn, error) {
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, nil)
	if err != nil {
		return nil, err
	}
	conn, err := dialer.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}

	return net.WrapConn(conn, true), nil
}
