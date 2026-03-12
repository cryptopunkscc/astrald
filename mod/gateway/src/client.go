package gateway

import (
	"errors"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/sig"
)

type connState uint8

const (
	connStateIdle     connState = iota
	connStateReserved connState = iota
	connStatePiped    connState = iota
)

type clientConn struct {
	exonet.Conn
	network string
	state   connState
	pipedTo *clientConn // non-nil when connStatePiped
}

type client struct {
	mu sync.Mutex

	Identity   *astral.Identity
	Nonce      astral.Nonce
	Visibility gateway.Visibility
	Target     *astral.Identity // nil for binders, set for clients
	//
	conns  sig.Set[*clientConn]
	pipeTo *clientConn // reserved binder conn for clients clients
}

func (c *client) isBinder() bool {
	return c.Target == nil
}

func (c *client) add(conn exonet.Conn) {
	c.conns.Add(&clientConn{
		Conn:    conn,
		network: conn.RemoteEndpoint().Network(),
		state:   connStateIdle,
	})
}

func (c *client) take() (*clientConn, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, cc := range c.conns.Clone() {
		if cc.state == connStateIdle {
			cc.state = connStateReserved
			return cc, true
		}
	}
	return nil, false
}

func (c *client) markPiped(cc, other *clientConn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cc.state = connStatePiped
	cc.pipedTo = other
}

func (c *client) takePipeTo() *clientConn {
	c.mu.Lock()
	defer c.mu.Unlock()
	cc := c.pipeTo
	c.pipeTo = nil
	return cc
}

func (c *client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	for _, cc := range c.conns.Clone() {
		errs = append(errs, cc.Close())
	}

	return errors.Join(errs...)
}

func (mod *Module) binderByIdentity(identity *astral.Identity) (*client, bool) {
	return mod.binders.Get(identity.String())
}

func (mod *Module) clientByNonce(nonce astral.Nonce) (*client, bool) {
	for _, c := range mod.binders.Values() {
		if c.Nonce == nonce {
			return c, true
		}
	}

	for _, c := range mod.clients.Clone() {
		if c.Nonce == nonce {
			return c, true
		}
	}
	return nil, false
}
