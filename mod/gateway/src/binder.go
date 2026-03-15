package gateway

import (
	"errors"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/sig"
)

// binder represents a node registered as reachable through the gateway.
// Only one binder registration per identity is allowed.
type binder struct {
	mu         sync.Mutex
	Identity   *astral.Identity
	Nonce      astral.Nonce
	Visibility gateway.Visibility
	conns      sig.Set[*bindingConn]
}

func (b *binder) addConn(conn exonet.Conn) *bindingConn {
	bc := newGatewayConn(conn, nil)
	bc.onClose = func() { b.conns.Remove(bc) }
	b.conns.Add(bc)
	return bc
}

// takeConn reserves an idle bindingConn for a connector via atomic CAS.
func (b *binder) takeConn() (*bindingConn, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, bc := range b.conns.Clone() {
		if bc.active.CompareAndSwap(false, true) {
			return bc, true
		}
	}
	return nil, false
}

func (b *binder) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	var errs []error
	for _, bc := range b.conns.Clone() {
		errs = append(errs, bc.Close())
	}
	return errors.Join(errs...)
}
