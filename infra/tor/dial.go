package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"golang.org/x/net/proxy"
)

const torProxyAddress = "127.0.0.1:9050"

func Dial(ctx context.Context, addr Addr) (infra.Conn, error) {
	torDialer, err := proxy.SOCKS5("tcp", torProxyAddress, nil, nil)
	if err != nil {
		return nil, err
	}

	conn, err := torDialer.Dial("tcp", addr.String()+":1791")
	if err != nil {
		return nil, err
	}

	return newConn(conn, addr, true), nil
}
