package tcp

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	_net "net"
)

func (mod *Module) Dial(ctx context.Context, endpoint exonet.Endpoint) (exonet.Conn, error) {
	switch endpoint.Network() {
	case "tcp", "inet":
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}

	var dialer = _net.Dialer{Timeout: mod.config.DialTimeout}

	tcpConn, err := dialer.DialContext(ctx, "tcp", endpoint.Address())
	if err != nil {
		return nil, err
	}

	return wrapTCPConn(tcpConn, true), nil
}
