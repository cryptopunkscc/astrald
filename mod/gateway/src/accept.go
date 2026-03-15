package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const (
	socketDeadTimeout      = 10 * time.Second
	socketProbeMaxAttempts = 3
	socketProbeTimeout     = 5 * time.Second
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
		go bc.keepalive()
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

	targetBinder, ok := mod.binderByIdentity(c.Target)
	if !ok {
		reserved.Close()
		conn.Close()
		return stopListener, nil
	}

	alive := mod.probeBinderConn(targetBinder, reserved)
	if alive == nil {
		mod.log.Errorv(1, "no alive conn for %v", c.Target)
		conn.Close()
		return stopListener, nil
	}

	cc := &connectorConn{
		Conn:    conn,
		network: conn.RemoteEndpoint().Network(),
		pipedTo: alive,
	}

	targetBinder.markPiped(alive, cc)
	mod.log.Infov(2, "pipe from %v to %v created", c.Identity, c.Target)
	go pipe(alive, cc)
	return stopListener, nil
}

// probeBinderConn signals a binderConn via its ping loop to verify liveness.
// It will try at most socketProbeMaxAttempts connections before giving up.
func (mod *Module) probeBinderConn(b *binder, reserved *binderConn) *binderConn {
	candidate := reserved

	for attempts := 0; attempts < socketProbeMaxAttempts; attempts++ {
		if candidate == nil {
			var ok bool
			candidate, ok = b.takeConn()
			if !ok {
				return nil
			}
		}

		if candidate.signal() {
			return candidate
		}
		candidate = nil
	}

	mod.log.Errorv(1, "binder %v probe exhausted", b.Identity)
	return nil
}
