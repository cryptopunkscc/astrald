package gateway

import (
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/mod/exonet"
)

var _ exonet.Conn = (*gatewayConn)(nil)

// note: maybe can be part of exonet
type deadliner interface {
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// gatewayConn wraps any io.ReadWriteCloser with gateway endpoint metadata.
type gatewayConn struct {
	io.ReadWriteCloser
	local    exonet.Endpoint
	remote   exonet.Endpoint
	outbound bool
}

func (c *gatewayConn) LocalEndpoint() exonet.Endpoint  { return c.local }
func (c *gatewayConn) RemoteEndpoint() exonet.Endpoint { return c.remote }
func (c *gatewayConn) Outbound() bool                  { return c.outbound }

func (c *gatewayConn) SetReadDeadline(t time.Time) error {
	if dl, ok := c.ReadWriteCloser.(deadliner); ok {
		return dl.SetReadDeadline(t)
	}
	return nil
}

func (c *gatewayConn) SetWriteDeadline(t time.Time) error {
	if dl, ok := c.ReadWriteCloser.(deadliner); ok {
		return dl.SetWriteDeadline(t)
	}
	return nil
}
