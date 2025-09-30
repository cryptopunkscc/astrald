package udp

import (
	"net"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/udp"
)

var _ exonet.Dialer = &Module{}

func (mod *Module) Dial(ctx *astral.Context, endpoint exonet.Endpoint) (exonet.Conn, error) {
	switch endpoint.Network() {
	case "udp":
		// Supported network
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}

	// Use net.Dialer for dialing UDP connections
	dialer := net.Dialer{Timeout: mod.config.DialTimeout}
	conn, err := dialer.DialContext(ctx, "udp", endpoint.Address())
	if err != nil {
		return nil, err
	}

	localEndpoint, _ := udp.ParseEndpoint(conn.LocalAddr().String())
	remoteEndpoint, _ := udp.ParseEndpoint(conn.RemoteAddr().String())

	udpConn, ok := conn.(*net.UDPConn)
	if !ok {
		return nil, exonet.ErrUnsupportedNetwork
	}

	reliableConn, err := NewConn(udpConn, localEndpoint, remoteEndpoint,
		mod.config.TransportConfig)
	if err != nil {
		return nil, err
	}

	reliableConn.outbound = true // mark as outbound connection

	return reliableConn, nil
}
