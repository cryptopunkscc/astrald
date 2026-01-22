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

	if mod.dial != nil {
		v, err := mod.dial.Get()
		if err != nil {
			return nil, err
		}

		if v != nil && !*v {
			return nil, exonet.ErrDisabledNetwork
		}
	}

	var dialer = _net.Dialer{Timeout: mod.config.DialTimeout}

	tcpConn, err := dialer.DialContext(ctx, "tcp", endpoint.Address())
	if err != nil {
		return nil, err
	}

	return tcp.WrapConn(tcpConn, true), nil
}
