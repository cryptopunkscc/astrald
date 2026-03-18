package gateway

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
)

// connector represents a pending connection request from a node that wants
// to reach a registered node through the gateway. Multiple connectors per withIdentity
// are allowed.
type connector struct {
	mu       sync.Mutex
	Identity *astral.Identity
	Nonce    astral.Nonce
	Target   *astral.Identity
	standby  *idleConn
}

// claimStandby atomically takes the standby idleConn, returning nil if
// already taken (connection already established or timed out).
func (c *connector) claimStandby() *idleConn {
	c.mu.Lock()
	defer c.mu.Unlock()

	bc := c.standby
	c.standby = nil
	return bc
}

func (c *connector) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.standby != nil {
		return c.standby.Close()
	}

	return nil
}
