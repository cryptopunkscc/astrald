package utp

import (
	"fmt"

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
		return nil, fmt.Errorf(`utp module/dial dialing endpoint failed: %w`,
			err)
	}

	localEndpoint, err := utpmod.ParseEndpoint(conn.LocalAddr().String())
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf(`utpmodule/dial parsing local endpoint failed
: %w`, err)
	}

	remoteEndpoint, err := utpmod.ParseEndpoint(conn.RemoteAddr().String())
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf(`utp module/dial parsing remote endpoint failed
: %w`, err)
	}

	return WrapUtpConn(conn, remoteEndpoint, localEndpoint, true), nil
}
