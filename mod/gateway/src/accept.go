package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const (
	socketDeadTimeout  = 10 * time.Second
	socketProbeTimeout = 5 * time.Second
)

// acceptSocketConn dispatches an incoming socket connection to either the binder
// or connector path based on the nonce it presents.
func (mod *Module) acceptSocketConn(_ context.Context, conn exonet.Conn) (stopListener bool, err error) {
	mod.log.Logv(2, "accepting socket connection from %v", conn.RemoteEndpoint())

	var nonce astral.Nonce
	if _, err := nonce.ReadFrom(conn); err != nil {
		mod.log.Errorv(1, "read nonce from %v: %v", conn.RemoteEndpoint(), err)
		conn.Close()
		return stopListener, nil
	}

	if b, ok := mod.binderByNonce(nonce); ok {
		mod.log.Infov(2, "added idle conn to binder %v", b.Identity)
		bc := b.addConn(conn)
		go bc.eventLoop(nil)
		return stopListener, nil
	}

	c, ok := mod.connectorByNonce(nonce)
	if !ok {
		mod.log.Errorv(1, "unknown nonce %v from %v", nonce, conn.RemoteEndpoint())
		conn.Close()
		return stopListener, nil
	}

	mod.connectors.Remove(c)

	reserved := c.takeReserved()
	if reserved == nil {
		conn.Close()
		return stopListener, fmt.Errorf("no reserved conn for %v", c.Target)
	}

	if !reserved.signal() {
		mod.log.Errorv(1, "reserved conn for %v is dead", c.Target)
		conn.Close()
		return stopListener, nil
	}

	mod.log.Infov(2, "pipe from %v to %v created", c.Identity, c.Target)
	go pipe(reserved, conn)
	return stopListener, nil
}
