package tcp

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	_net "net"
)

func (mod *Module) Dial(ctx context.Context, endpoint net.Endpoint) (net.Conn, error) {
	switch endpoint.Network() {
	case "tcp", "inet":
	default:
		return nil, infra.ErrUnsupportedNetwork
	}

	var dialer = _net.Dialer{Timeout: mod.config.DialTimeout}

	tcpConn, err := dialer.DialContext(ctx, "tcp", endpoint.String())
	if err != nil {
		return nil, err
	}

	return wrapTCPConn(tcpConn, true), nil
}
