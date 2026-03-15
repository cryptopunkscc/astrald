package gateway

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

// deadliner is implemented by connections that support read/write deadlines (e.g. net.Conn).
type deadliner interface {
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// bindingConn is a unified idle socket connection for both binder and gateway sides
type bindingConn struct {
	exonet.Conn
	closed   atomic.Bool
	active   atomic.Bool     // set on idle→active; guards idle counters against double-decrement
	dead     chan struct{}   // closed when keepalive exits
	signalCh chan chan error // non-nil → gateway (responder) mode
	onClose  func()
}

func newGatewayConn(conn exonet.Conn, onClose func()) *bindingConn {
	return &bindingConn{
		Conn:     conn,
		dead:     make(chan struct{}),
		signalCh: make(chan chan error, 1),
		onClose:  onClose,
	}
}

func newBinderConn(conn exonet.Conn, onClose func()) *bindingConn {
	return &bindingConn{
		Conn:    conn,
		dead:    make(chan struct{}),
		onClose: onClose,
	}
}

func (bc *bindingConn) Close() error {
	err := bc.Conn.Close()
	if !bc.closed.Swap(true) && bc.onClose != nil {
		bc.onClose()
	}
	return err
}

// keepalive runs the ping/pong loop until activation or connection loss.
func (bc *bindingConn) keepalive(done <-chan struct{}, onActive func(), onActivate func() error) {
	defer close(bc.dead)
	activated := false
	defer func() {
		if !activated {
			bc.Close()
		}
	}()

	binder := bc.signalCh == nil
	dl, _ := bc.Conn.(deadliner)

	for {
		if binder {
			if err := gateway.WritePing(bc.Conn); err != nil {
				return
			}
		}

		timeout := socketDeadTimeout
		if binder {
			timeout = socketPingTimeout
		}
		if dl != nil {
			dl.SetReadDeadline(time.Now().Add(timeout))
		}
		var b [1]byte
		_, err := io.ReadFull(bc.Conn, b[:])
		if dl != nil {
			dl.SetReadDeadline(time.Time{})
		}
		if err != nil {
			return
		}

		switch b[0] {
		case gateway.BytePing: // gateway only
			select {
			case respCh := <-bc.signalCh:
				bc.sendSignalGo(respCh)
				activated = true
				return
			default:
				if err := gateway.WritePong(bc.Conn); err != nil {
					return
				}
			}
		case gateway.BytePong: // binder only
			select {
			case <-time.After(socketPingInterval):
			case <-done:
				return
			}
		case gateway.ByteSignalGo: // binder only
			// Set active before onActive so that if Close fires during WriteSignalReady,
			// onClose sees active=true and does not double-decrement the idle counter.
			bc.active.Store(true)
			if onActive != nil {
				onActive()
			}
			if err := gateway.WriteSignalReady(bc.Conn); err != nil {
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
	dl, _ := bc.Conn.(deadliner)
	if dl != nil {
		defer dl.SetReadDeadline(time.Time{})
		defer dl.SetWriteDeadline(time.Time{})
		dl.SetWriteDeadline(time.Now().Add(socketProbeTimeout))
	}
	if err := gateway.WriteSignalGo(bc.Conn); err != nil {
		respCh <- err
		return
	}
	if dl != nil {
		dl.SetReadDeadline(time.Now().Add(socketProbeTimeout))
	}
	var b [1]byte
	if _, err := io.ReadFull(bc.Conn, b[:]); err != nil {
		respCh <- err
		return
	}
	if b[0] != gateway.ByteSignalReady {
		respCh <- fmt.Errorf("expected signalReady, got 0x%02x", b[0])
		return
	}
	respCh <- nil
}

// signal queues a ByteSignalGo via the keepalive loop and waits for acknowledgement.
// Returns true if the binder acknowledged. Gateway mode only.
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
