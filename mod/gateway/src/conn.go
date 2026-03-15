package gateway

import (
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/mod/exonet"
)

var _ exonet.Conn = (*gwConn)(nil)

// note: maybe can be part of exonet
type deadliner interface {
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// gwConn wraps any io.ReadWriteCloser with gateway endpoint metadata.
type gwConn struct {
	io.ReadWriteCloser
	local    exonet.Endpoint
	remote   exonet.Endpoint
	outbound bool
}

func (c *gwConn) LocalEndpoint() exonet.Endpoint  { return c.local }
func (c *gwConn) RemoteEndpoint() exonet.Endpoint { return c.remote }
func (c *gwConn) Outbound() bool                  { return c.outbound }

func (c *gwConn) SetReadDeadline(t time.Time) error {
	if dl, ok := c.ReadWriteCloser.(deadliner); ok {
		return dl.SetReadDeadline(t)
	}
	return nil
}

func (c *gwConn) SetWriteDeadline(t time.Time) error {
	if dl, ok := c.ReadWriteCloser.(deadliner); ok {
		return dl.SetWriteDeadline(t)
	}
	return nil
}
