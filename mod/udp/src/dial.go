package udp

import (
	"net"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/udp"
	"github.com/cryptopunkscc/astrald/mod/udp/rudp"
)

var _ exonet.Dialer = &Module{}

// Dial establishes a reliable (rudp) connection and returns it only after the
// RUDP client handshake succeeds. Timeout / cancellation behavior:
//   - If the caller's context has a deadline, it governs both dial and handshake.
//   - Otherwise net.Dialer.Timeout (DialTimeout in config, if >0) limits only the dial phase.
//   - The handshake then reuses the original context (may block if no deadline provided).
func (mod *Module) Dial(ctx *astral.Context, endpoint exonet.Endpoint) (exonet.Conn, error) {
	switch endpoint.Network() {
	case "udp":
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}

	dialer := net.Dialer{}
	if mod.config.DialTimeout > 0 {
		dialer.Timeout = mod.config.DialTimeout
	}

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

	reliableConn, err := rudp.NewConn(udpConn, localEndpoint, remoteEndpoint,
		mod.config.TransportConfig, true, nil, ctx)
	if err != nil {
		return nil, err
	}

	err = reliableConn.StartClientHandshake(ctx)
	if err != nil {
		reliableConn.Close()
		return nil, err
	}

	return reliableConn, nil
}
