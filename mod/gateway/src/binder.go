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

// binderConn is a connection pre-opened by a binder node to the gateway,
// sitting idle until a connector claims it.
type binderConn struct {
	exonet.Conn
	state   connState
	pipedTo *connectorConn
	onClose func()
}

func (bc *binderConn) Close() error {
	err := bc.Conn.Close()
	if bc.onClose != nil {
		bc.onClose()
	}
	return err
}

// binder represents a node registered as reachable through the gateway.
// Only one binder registration per identity is allowed.
type binder struct {
	mu         sync.Mutex
	Identity   *astral.Identity
	Nonce      astral.Nonce
	Visibility gateway.Visibility
	conns      sig.Set[*binderConn]
}

func (b *binder) addConn(conn exonet.Conn) {
	bc := &binderConn{
		Conn:  conn,
		state: connStateIdle,
	}
	bc.onClose = func() { b.conns.Remove(bc) }
	b.conns.Add(bc)
}

// takeConn reserves an idle binderConn for a connector.
func (b *binder) takeConn() (*binderConn, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, bc := range b.conns.Clone() {
		if bc.state == connStateIdle {
			bc.state = connStateReserved
			return bc, true
		}
	}
	return nil, false
}

func (b *binder) markPiped(bc *binderConn, cc *connectorConn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	bc.state = connStatePiped
	bc.pipedTo = cc
}

func (b *binder) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	var errs []error
	for _, bc := range b.conns.Clone() {
		if bc.state == connStateReserved {
			continue
		}
		errs = append(errs, bc.Close())
	}
	return errors.Join(errs...)
}
