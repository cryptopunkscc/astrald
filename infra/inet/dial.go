package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"net"
)

func Dial(_ context.Context, addr Addr) (infra.Conn, error) {
	tcpConn, err := net.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}

	return newConn(tcpConn, true), nil
}
