package tcp

import (
	_net "net"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

func (mod *Module) Dial(ctx *astral.Context, endpoint exonet.Endpoint) (exonet.Conn, error) {
	switch endpoint.Network() {
	case "tcp", "inet":
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}

	if v, _ := mod.dial.Value(); v == nil || !*v {
		return nil, exonet.ErrUnsupportedNetwork
	}

	var dialer = _net.Dialer{Timeout: mod.config.DialTimeout}

	tcpConn, err := dialer.DialContext(ctx, "tcp", endpoint.Address())
	if err != nil {
		return nil, err
	}

	return tcp.WrapConn(tcpConn, true), nil
}
