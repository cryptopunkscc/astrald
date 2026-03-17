package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/sig"
)

// registeredNode represents a node registered as reachable through the gateway.
// Only one registration per identity is allowed.
type registeredNode struct {
	Identity   *astral.Identity
	Nonce      astral.Nonce
	Visibility gateway.Visibility
	idleConns  sig.Set[*standbyConn]
}

func (b *registeredNode) registerConn(conn exonet.Conn) *standbyConn {
	bc := newGatewayConn(conn, roleGateway)
	b.idleConns.Add(bc)
	go func() { <-bc.doneCh; b.idleConns.Remove(bc) }()
	return bc
}

// claimConn reserves an idle standbyConn for a connector via atomic CAS.
func (b *registeredNode) claimConn() (*standbyConn, bool) {
	for _, c := range b.idleConns.Clone() {
		if c.claimed.CompareAndSwap(false, true) {
			return c, true
		}
	}
	return nil, false
}

func (b *registeredNode) Close() error {
	for _, c := range b.idleConns.Clone() {
		c.Close()
	}
	return nil
}
