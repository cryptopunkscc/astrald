package gateway

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

// connectorConn is the connection opened by a connector node to the gateway,
// to be piped to a reserved binderConn.
type connectorConn struct {
	exonet.Conn
	network string
	pipedTo *binderConn
}

// connector represents a pending connection request from a node that wants
// to reach a binder through the gateway. Multiple connectors per identity
// are allowed.
type connector struct {
	mu       sync.Mutex
	Identity *astral.Identity
	Nonce    astral.Nonce
	Target   *astral.Identity
	reserved *binderConn
}

// takeReserved atomically takes the reserved binderConn, returning nil if
// already taken (connection already established or timed out).
func (c *connector) takeReserved() *binderConn {
	c.mu.Lock()
	defer c.mu.Unlock()

	bc := c.reserved
	c.reserved = nil
	return bc
}

func (c *connector) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.reserved != nil {
		return c.reserved.Close()
	}

	return nil
}
