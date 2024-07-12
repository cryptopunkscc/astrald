package tcp

import (
	"context"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/net"
	_net "net"
)

func (mod *Module) Dial(ctx context.Context, endpoint net.Endpoint) (net.Conn, error) {
	switch endpoint.Network() {
	case "tcp", "inet":
	default:
		return nil, core.ErrUnsupportedNetwork
	}

	var dialer = _net.Dialer{Timeout: mod.config.DialTimeout}

	tcpConn, err := dialer.DialContext(ctx, "tcp", endpoint.String())
	if err != nil {
		return nil, err
	}

	return wrapTCPConn(tcpConn, true), nil
}
