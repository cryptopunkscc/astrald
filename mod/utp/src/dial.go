package utp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	utpmod "github.com/cryptopunkscc/astrald/mod/utp"
	"github.com/cryptopunkscc/utp"
)

var _ exonet.Dialer = &Module{}

// Dial establishes a reliable (utp) connection and wraps it as an exonet.Conn.
func (mod *Module) Dial(ctx *astral.Context, endpoint exonet.Endpoint) (exonet.Conn, error) {
	switch endpoint.Network() {
	case "utp":
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}

	dialer := utp.Dialer{Timeout: mod.config.DialTimeout}
	conn, err := dialer.Dial("utp", endpoint.Address())
	if err != nil {
		return nil, err
	}

	localEndpoint, err := utpmod.ParseEndpoint(conn.LocalAddr().String())
	if err != nil {
		conn.Close()
		return nil, err
	}

	remoteEndpoint, err := utpmod.ParseEndpoint(conn.RemoteAddr().String())
	if err != nil {
		conn.Close()
		return nil, err
	}

	return WrapUtpConn(conn, remoteEndpoint, localEndpoint, true), nil
}
