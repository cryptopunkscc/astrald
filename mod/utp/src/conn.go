package utp

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/utp"
)

type WrappedConn struct {
	outbound bool
	renote   exonet.Endpoint
	local    exonet.Endpoint

	*utp.Conn
}

func (w WrappedConn) Outbound() bool {
	return w.outbound
}

func (w WrappedConn) LocalEndpoint() exonet.Endpoint {
	return w.local
}

func (w WrappedConn) RemoteEndpoint() exonet.Endpoint {
	return w.renote
}

func WrapUtpConn(
	conn *utp.Conn,
	remote exonet.Endpoint,
	local exonet.Endpoint,
	outbound bool) exonet.Conn {
	return WrappedConn{
		outbound: outbound,
		Conn:     conn,
		renote:   remote,
		local:    local,
	}
}
