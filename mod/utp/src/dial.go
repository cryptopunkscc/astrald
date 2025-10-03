package utp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	utpmod "github.com/cryptopunkscc/astrald/mod/utp"
	"github.com/cryptopunkscc/utp"
)

var _ exonet.Dialer = &Module{}

// Dial establishes a reliable (utp) connection and returns it only after the
func (mod *Module) Dial(ctx *astral.Context, endpoint exonet.Endpoint) (exonet.Conn, error) {

	switch endpoint.Network() {
	case "utp":
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}

	dialer := utp.Dialer{Timeout: mod.config.DialTimeout}
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
