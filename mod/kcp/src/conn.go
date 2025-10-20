package kcp

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
	kcpgo "github.com/xtaci/kcp-go/v5"
)

type WrappedConn struct {
	*kcpgo.UDPSession
	remote   exonet.Endpoint
	local    exonet.Endpoint
	outbound bool
}

func (w WrappedConn) Outbound() bool {
	return w.outbound
}

func (w WrappedConn) LocalEndpoint() exonet.Endpoint {
	return w.local
}

func (w WrappedConn) RemoteEndpoint() exonet.Endpoint {
	return w.remote
}

func WrapKCPConn(
	sess *kcpgo.UDPSession,
	remote exonet.Endpoint,
	local exonet.Endpoint,
	outbound bool) exonet.Conn {
	return WrappedConn{
		outbound:   outbound,
		UDPSession: sess,
		remote:     remote,
		local:      local,
	}
}
