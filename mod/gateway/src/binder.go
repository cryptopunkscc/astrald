package gateway

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/gateway/src/frames"
	"github.com/cryptopunkscc/astrald/sig"
)

// binder represents a node registered as reachable through the gateway.
// Only one binder registration per identity is allowed.
type binder struct {
	Identity   *astral.Identity
	Nonce      astral.Nonce
	Visibility gateway.Visibility
	conns      sig.Set[*bindingConn]
}

func (b *binder) addConn(conn exonet.Conn) *bindingConn {
	var bc *bindingConn
	bc = newGatewayConn(conn, func() { b.conns.Remove(bc) })
	b.conns.Add(bc)
	return bc
}

// takeConn reserves an idle bindingConn for a connector via atomic CAS.
// CAS on active is sufficient — no mutex needed.
func (b *binder) takeConn() (*bindingConn, bool) {
	for _, bc := range b.conns.Clone() {
		if bc.active.CompareAndSwap(false, true) {
			return bc, true
		}
	}
	return nil, false
}

func (b *binder) Close() error {
	for _, bc := range b.conns.Clone() {
		bc.Close()
	}
	return nil
}

type connRole uint8

const (
	roleBinder connRole = iota
	roleGateway
)

// bindingConn is a unified idle socket connection for both binder and gateway sides
type bindingConn struct {
	exonet.Conn
	role connRole

	closed atomic.Bool
	active atomic.Bool // set on idle→active; guards idle counters against double-decrement

	dead     chan struct{}
	signalCh chan chan error
	onClose  func()
}

func newGatewayConn(conn exonet.Conn, onClose func()) *bindingConn {
	return &bindingConn{
		Conn:     conn,
		role:     roleGateway,
		dead:     make(chan struct{}),
		signalCh: make(chan chan error, 1),
		onClose:  onClose,
	}
}

func newBinderConn(conn exonet.Conn) *bindingConn {
	return &bindingConn{
		Conn: conn,
		role: roleBinder,
		dead: make(chan struct{}),
	}
}

func (bc *bindingConn) SetReadDeadline(t time.Time) error {
	if dl, ok := bc.Conn.(deadliner); ok {
		return dl.SetReadDeadline(t)
	}
	return nil
}

func (bc *bindingConn) SetWriteDeadline(t time.Time) error {
	if dl, ok := bc.Conn.(deadliner); ok {
		return dl.SetWriteDeadline(t)
	}
	return nil
}

// readFrame reads a single control byte within the gateway keepalive/control phase.
// Must not be called after activation, when the connection becomes a raw stream.
func (bc *bindingConn) readFrame(timeout time.Duration) (byte, error) {
	bc.SetReadDeadline(time.Now().Add(timeout))
	var b [1]byte
	_, err := io.ReadFull(bc.Conn, b[:])
	bc.SetReadDeadline(time.Time{})
	return b[0], err
}

func (bc *bindingConn) Close() error {
	err := bc.Conn.Close()
	if !bc.closed.Swap(true) && bc.onClose != nil {
		bc.onClose()
	}
	return err
}

// keepalive runs the ping/pong loop until activation or connection loss.
// done stops the binder-side ping sleep on shutdown (pass ctx.Done()).
// onActivate is called after WriteSignalReady; returned error causes bc to be closed.
func (bc *bindingConn) keepalive(done <-chan struct{}, onActivate func() error) {
	defer close(bc.dead)
	activated := false
	defer func() {
		if !activated {
			bc.Close()
		}
	}()

	for {
		if bc.role == roleBinder {
			if err := frames.WritePing(bc.Conn); err != nil {
				return
			}
		}

		timeout := socketDeadTimeout
		if bc.role == roleBinder {
			timeout = socketPingTimeout
		}
		frame, err := bc.readFrame(timeout)
		if err != nil {
			return
		}

		switch frame {
		case frames.BytePing: // roleGateway only
			select {
			case respCh := <-bc.signalCh:
				bc.sendSignalGo(respCh)
				activated = true
				return
			default:
				if err := frames.WritePong(bc.Conn); err != nil {
					return
				}
			}
		case frames.BytePong: // roleBinder only
			select {
			case <-time.After(socketPingInterval):
			case <-done:
				return
			}
		case frames.ByteSignalGo: // roleBinder only
			bc.active.Store(true)
			if err := frames.WriteSignalReady(bc.Conn); err != nil {
				bc.Close()
				activated = true // active is set; defer must not double-close
				return
			}
			if onActivate != nil {
				if err := onActivate(); err != nil {
					bc.Close()
				}
			}
			activated = true
			return
		default:
			return
		}
	}
}

// sendSignalGo sends ByteSignalGo and waits for ByteSignalReady, reporting the result on respCh.
func (bc *bindingConn) sendSignalGo(respCh chan error) {
	defer bc.SetWriteDeadline(time.Time{})
	bc.SetWriteDeadline(time.Now().Add(socketProbeTimeout))

	if err := frames.WriteSignalGo(bc.Conn); err != nil {
		respCh <- err
		return
	}

	frame, err := bc.readFrame(socketProbeTimeout)
	if err != nil {
		respCh <- err
		return
	}
	if frame != frames.ByteSignalReady {
		respCh <- fmt.Errorf("expected signalReady, got 0x%02x", frame)
		return
	}
	respCh <- nil
}

func (bc *bindingConn) signal() bool {
	respCh := make(chan error, 1)
	select {
	case bc.signalCh <- respCh:
	case <-time.After(socketProbeTimeout):
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
		return false
	case <-time.After(socketProbeTimeout + socketPingInterval):
		bc.Close()
		return false
	}
}
