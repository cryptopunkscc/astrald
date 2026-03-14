package gateway

import (
	"io"

	"github.com/cryptopunkscc/astrald/mod/exonet"
)

var _ exonet.Conn = (*gwConn)(nil)

type gwConn struct {
	io.ReadWriteCloser
	local    exonet.Endpoint
	remote   exonet.Endpoint
	outbound bool
}

func (c *gwConn) LocalEndpoint() exonet.Endpoint  { return c.local }
func (c *gwConn) RemoteEndpoint() exonet.Endpoint { return c.remote }
func (c gwConn) Outbound() bool                   { return c.outbound }
