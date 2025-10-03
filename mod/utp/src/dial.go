package utp

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	utpmod "github.com/cryptopunkscc/astrald/mod/utp"
	"github.com/cryptopunkscc/utp"
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

	laddr, err := utp.ResolveAddr("utp", endpoint.Address())
	if err != nil {
		return nil, fmt.Errorf("Dial failed to resolve address %s",
			endpoint.Address())
	}

	dialer := utp.Dialer{LocalAddr: laddr, Timeout: mod.config.DialTimeout}
	if mod.config.DialTimeout > 0 {
		dialer.Timeout = mod.config.DialTimeout
	}

	conn, err := dialer.Dial("utp", endpoint.Address())
	if err != nil {
		return nil, err
	}

	localEndpoint, _ := utpmod.ParseEndpoint(conn.LocalAddr().String())
	remoteEndpoint, _ := utpmod.ParseEndpoint(conn.RemoteAddr().String())

	return WrapUtpConn(conn, remoteEndpoint, localEndpoint, true), nil
}
