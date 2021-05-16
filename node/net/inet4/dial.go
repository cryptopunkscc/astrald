package inet4

import (
	"context"
	net2 "github.com/cryptopunkscc/astrald/node/net"
	_net "net"
)

func Dial(_ context.Context, ep net2.Endpoint) (net2.Conn, error) {
	tcpConn, err := _net.Dial("tcp4", ep.Address)
	if err != nil {
		return nil, err
	}

	return &Conn{
		Conn:     tcpConn,
		outbound: true,
	}, nil
}
