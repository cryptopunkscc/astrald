package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"net"
	"time"
)

const dialTimeout = 10 * time.Second

var dialConfig = net.Dialer{Timeout: dialTimeout}

func (inet *Inet) Dial(ctx context.Context, addr Addr) (infra.Conn, error) {
	tcpConn, err := dialConfig.DialContext(ctx, "tcp", addr.String())
	if err != nil {
		return nil, err
	}

	return newConn(tcpConn, true), nil
}

func (inet *Inet) DialFrom(ctx context.Context, addr Addr, from Addr) (infra.Conn, error) {
	var err error
	var config = dialConfig

	config.LocalAddr, err = net.ResolveTCPAddr("tcp", from.String())
	if err != nil {
		return nil, err
	}

	tcpConn, err := config.DialContext(ctx, "tcp", addr.String())
	if err != nil {
		return nil, err
	}

	return newConn(tcpConn, true), nil
}
