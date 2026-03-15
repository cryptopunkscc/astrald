package gateway

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/sig"
)

// deadliner is implemented by connections that support read/write deadlines (e.g. net.Conn).
type deadliner interface {
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

func withReadDeadline(conn exonet.Conn, d time.Duration) {
	if dl, ok := conn.(deadliner); ok {
		dl.SetReadDeadline(time.Now().Add(d))
	}
}

func withWriteDeadline(conn exonet.Conn, d time.Duration) {
	if dl, ok := conn.(deadliner); ok {
		dl.SetWriteDeadline(time.Now().Add(d))
	}
}

func clearDeadlines(conn exonet.Conn) {
	if dl, ok := conn.(deadliner); ok {
		dl.SetReadDeadline(time.Time{})
		dl.SetWriteDeadline(time.Time{})
	}
}

type connState uint8

const (
	stateIdle     connState = iota
	stateReserved connState = iota
	statePiped    connState = iota
)

// binderConn is a connection pre-opened by a binder node to the gateway,
// sitting idle until a connector claims it.
type binderConn struct {
	exonet.Conn
	state    connState
	pipedTo  *connectorConn
	onClose  func()
	signalCh chan chan error // connector requests keepalive to send ByteSignalGo
	dead     chan struct{}   // closed when the connection is no longer alive
	closed   atomic.Bool
}

func (bc *binderConn) IsIdle() bool     { return bc.state == stateIdle }
func (bc *binderConn) IsReserved() bool { return bc.state == stateReserved }
func (bc *binderConn) IsPiped() bool    { return bc.state == statePiped }

func (bc *binderConn) Close() error {
	err := bc.Conn.Close()
	if !bc.closed.Swap(true) && bc.onClose != nil {
		bc.onClose()
	}
	return err
}

// keepalive reads binder pings and responds with pong.
// When a connector signals via signalCh it sends ByteSignalGo and waits for ByteSignalReady.
func (bc *binderConn) keepalive() {
	defer close(bc.dead)
	// piped tracks whether pipe() has taken ownership of bc.
	// If true, pipe() is responsible for closing — keepalive must not.
	piped := false
	defer func() {
		if !piped {
			bc.Close()
		}
	}()

	for {
		withReadDeadline(bc.Conn, socketDeadTimeout)
		var b [1]byte
		if _, err := io.ReadFull(bc.Conn, b[:]); err != nil {
			return
		}
		clearDeadlines(bc.Conn)

		switch b[0] {
		case gateway.BytePing:
			select {
			case respCh := <-bc.signalCh:
				piped = true // pipe or probeBinderConn takes ownership; don't close on exit
				bc.sendSignal(respCh)
				return
			default:
				if err := gateway.WritePong(bc.Conn); err != nil {
					return
				}
			}
		default:
			return
		}
	}
}

// sendSignal sends ByteSignalGo to the binder and waits for ByteSignalReady.
// Called from keepalive when a connector signals via signalCh.
func (bc *binderConn) sendSignal(respCh chan error) {
	withWriteDeadline(bc.Conn, socketProbeTimeout)
	if err := gateway.WriteSignalGo(bc.Conn); err != nil {
		respCh <- err
		return
	}
	clearDeadlines(bc.Conn)
	withReadDeadline(bc.Conn, socketProbeTimeout)
	var b [1]byte
	if _, err := io.ReadFull(bc.Conn, b[:]); err != nil {
		respCh <- err
		return
	}
	clearDeadlines(bc.Conn)
	if b[0] != gateway.ByteSignalReady {
		respCh <- fmt.Errorf("expected signalReady, got 0x%02x", b[0])
		return
	}
	respCh <- nil
}

// signal queues a ByteSignalGo request via the keepalive loop and waits for the result.
// Returns true if the conn is alive and was signaled successfully.
// Closes bc on failure unless keepalive already exited.
func (bc *binderConn) signal() bool {
	respCh := make(chan error, 1)
	select {
	case bc.signalCh <- respCh:
	case <-time.After(socketProbeTimeout):
		// signalCh buffer full — another connector is already signaling this conn
		bc.Close()
		return false
	}

	select {
	case err := <-respCh:
		if err != nil {
			bc.Close()
		}
		return err == nil
	case <-bc.dead:
		// keepalive exited — conn was dead
		return false
	case <-time.After(socketProbeTimeout + socketPingInterval):
		bc.Close()
		return false
	}
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

func (b *binder) addConn(conn exonet.Conn) *binderConn {
	bc := &binderConn{
		Conn:     conn,
		state:    stateIdle,
		signalCh: make(chan chan error, 1),
		dead:     make(chan struct{}),
	}
	bc.onClose = func() { b.conns.Remove(bc) }
	b.conns.Add(bc)
	return bc
}

// takeConn reserves an idle binderConn for a connector.
func (b *binder) takeConn() (*binderConn, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, bc := range b.conns.Clone() {
		if bc.IsIdle() {
			bc.state = stateReserved
			return bc, true
		}
	}
	return nil, false
}

func (b *binder) markPiped(bc *binderConn, cc *connectorConn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	bc.state = statePiped
	bc.pipedTo = cc
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
