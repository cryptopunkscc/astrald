package tcp

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/net"
	_net "net"
)

func (drv *driver) Dial(_ context.Context, addr net.Addr) (net.Conn, error) {
	tcpConn, err := _net.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}

	return net.WrapConn(tcpConn, true), nil
}
