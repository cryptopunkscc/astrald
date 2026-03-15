package gateway

import (
	"io"

	"github.com/cryptopunkscc/astrald/mod/exonet"
)

var _ exonet.Conn = (*gwConn)(nil)

// routeConn adapts an io.ReadWriteCloser (e.g. astral.Conn from query.Route) to exonet.Conn.
// Endpoint methods return nil — gwConn overrides all of them.
type routeConn struct {
	io.ReadWriteCloser
}

func (c *routeConn) LocalEndpoint() exonet.Endpoint  { return nil }
func (c *routeConn) RemoteEndpoint() exonet.Endpoint { return nil }
func (c *routeConn) Outbound() bool                  { return true }

type gwConn struct {
	*bindingConn
	local    exonet.Endpoint
	remote   exonet.Endpoint
	outbound bool
}

func (c *gwConn) LocalEndpoint() exonet.Endpoint  { return c.local }
func (c *gwConn) RemoteEndpoint() exonet.Endpoint { return c.remote }
func (c gwConn) Outbound() bool                   { return c.outbound }
