package gateway

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
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
	bc := newGatewayConn(conn)
	b.conns.Add(bc)
	go func() { <-bc.Closed(); b.conns.Remove(bc) }()
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

// bindingConn is a unified idle socket connection for both binder and gateway sides.
type bindingConn struct {
	exonet.Conn
	role connRole

	closed    atomic.Bool
	active    atomic.Bool // set idle→reserved by takeConn CAS; also stops binder pings
	handedOff atomic.Bool // set when conn is handed off to higher-level; suppresses Close in defers
	deadOnce  sync.Once

	dead          chan struct{}
	signalCh      chan chan error    // gateway-side: receives signal requests from connector
	writeCh       chan astral.Object // readLoop→writeLoop: outbound frames
	activatedCh   chan struct{}      // closed when activation completes
	closedCh      chan struct{}      // closed when Close() is called
	activatedOnce sync.Once
}

func newGatewayConn(conn exonet.Conn) *bindingConn {
	return &bindingConn{
		Conn:        conn,
		role:        roleGateway,
		dead:        make(chan struct{}),
		signalCh:    make(chan chan error, 1),
		writeCh:     make(chan astral.Object, 4),
		activatedCh: make(chan struct{}),
		closedCh:    make(chan struct{}),
	}
}

func newBinderConn(conn exonet.Conn) *bindingConn {
	return &bindingConn{
		Conn:        conn,
		role:        roleBinder,
		dead:        make(chan struct{}),
		writeCh:     make(chan astral.Object, 4),
		activatedCh: make(chan struct{}),
		closedCh:    make(chan struct{}),
	}
}

// Activated returns a channel that is closed when the conn is handed off to
// higher-level code (after the last frame is flushed on the binder side, or
// after the signal handshake completes on the gateway side).
func (bc *bindingConn) Activated() <-chan struct{} { return bc.activatedCh }

// Closed returns a channel that is closed when the underlying connection is closed.
func (bc *bindingConn) Closed() <-chan struct{} { return bc.closedCh }

func (bc *bindingConn) setActivated() {
	bc.activatedOnce.Do(func() { close(bc.activatedCh) })
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

// Close closes the underlying connection exactly once.
func (bc *bindingConn) Close() error {
	if bc.closed.Swap(true) {
		return nil
	}
	err := bc.Conn.Close()
	close(bc.closedCh)
	return err
}

func (bc *bindingConn) closeDead() {
	bc.deadOnce.Do(func() { close(bc.dead) })
}

// eventLoop starts the read and write loops.
func (bc *bindingConn) eventLoop(done <-chan struct{}) {
	ch := channel.New(bc)
	go bc.writeLoop(ch, done)
	bc.readLoop(ch)
}

// readLoop owns all incoming frame handling via ch.Switch.
// On exit it closes writeCh to signal writeLoop to flush and stop.
func (bc *bindingConn) readLoop(ch *channel.Channel) {
	defer bc.closeDead()
	defer close(bc.writeCh)
	defer func() {
		if !bc.handedOff.Load() {
			bc.Close()
		}
	}()

	if bc.role == roleBinder {
		// Safety net: writeLoop sets socketPingTimeout after each ping, but we need
		// a deadline in place from the start in case writeLoop hasn't run yet.
		bc.SetReadDeadline(time.Now().Add(socketDeadTimeout))
		ch.Switch(
			func(*frames.PongMsg) error {
				return nil
			},
			func(*frames.SignalGoMsg) error {
				bc.active.Store(true) // stops writeLoop ping ticker
				select {
				case bc.writeCh <- &frames.SignalReadyMsg{}:
				default:
					return fmt.Errorf("write buffer full")
				}
				bc.handedOff.Store(true) // only after successful enqueue
				return channel.ErrBreak
			},
		)
		return
	}

	// roleGateway
	var pendingRespCh chan error
	bc.SetReadDeadline(time.Now().Add(socketDeadTimeout))
	ch.Switch(
		func(*frames.PingMsg) error {
			bc.SetReadDeadline(time.Now().Add(socketDeadTimeout))
			select {
			case respCh := <-bc.signalCh:
				pendingRespCh = respCh
				select {
				case bc.writeCh <- &frames.SignalGoMsg{}:
				default:
					return fmt.Errorf("write buffer full")
				}
				return nil
			default:
				select {
				case bc.writeCh <- &frames.PongMsg{}:
				default:
					return fmt.Errorf("write buffer full")
				}
				return nil
			}
		},
		func(*frames.SignalReadyMsg) error {
			if pendingRespCh == nil {
				return fmt.Errorf("unexpected SignalReady")
			}
			pendingRespCh <- nil
			bc.handedOff.Store(true)
			bc.setActivated()
			return channel.ErrBreak
		},
	)
}

// writeLoop owns all outbound traffic.
// For roleBinder it drives the ping ticker and fires Activated() after flushing writeCh.
// For roleGateway it only drains writeCh.
func (bc *bindingConn) writeLoop(ch *channel.Channel, done <-chan struct{}) {
	defer bc.closeDead()
	defer func() {
		if !bc.handedOff.Load() {
			bc.Close()
		}
	}()

	if bc.role == roleBinder {
		if err := ch.Send(&frames.PingMsg{}); err != nil {
			return
		}
		bc.SetReadDeadline(time.Now().Add(socketPingTimeout))

		ticker := time.NewTicker(socketPingInterval)
		defer ticker.Stop()
		tickCh := ticker.C

		for {
			select {
			case <-done:
				return
			case obj, ok := <-bc.writeCh:
				if !ok {
					if bc.handedOff.Load() {
						bc.setActivated() // fired after SignalReadyMsg is flushed
					}
					return
				}
				if err := ch.Send(obj); err != nil {
					return
				}
			case <-tickCh:
				if bc.active.Load() {
					ticker.Stop()
					tickCh = nil
					continue
				}
				if err := ch.Send(&frames.PingMsg{}); err != nil {
					return
				}
				bc.SetReadDeadline(time.Now().Add(socketPingTimeout))
			}
		}
	}

	// roleGateway: no pings, just drain writeCh
	for obj := range bc.writeCh {
		if err := ch.Send(obj); err != nil {
			return
		}
	}
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
