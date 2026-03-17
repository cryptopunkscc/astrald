package gateway

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

var ErrConnClosed = errors.New("conn closed")

type connRole uint8

const (
	roleClient connRole = iota
	roleGateway
)

func (r connRole) String() string {
	switch r {
	case roleClient:
		return "client"
	case roleGateway:
		return "gateway"
	default:
		return "unknown"
	}
}

// standbyConn is a unified idle socket connection for both node and gateway sides.
type standbyConn struct {
	exonet.Conn
	role connRole
	log  *log.Logger

	closed   atomic.Bool
	claimed  atomic.Bool
	relaying atomic.Bool

	handoffCh   chan struct{}
	handoffOnce sync.Once
	readyCh     chan struct{}
	doneCh      chan struct{}
	readyOnce   sync.Once
}

func newGatewayConn(conn exonet.Conn, role connRole, l *log.Logger) *standbyConn {
	bc := &standbyConn{
		Conn:    conn,
		role:    role,
		log:     l,
		readyCh: make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
	if role == roleGateway {
		bc.handoffCh = make(chan struct{})
	}
	l.Logv(2, "standby conn created role=%v remote=%v", role, conn.RemoteEndpoint())
	return bc
}

func (conn *standbyConn) markReady() {
	conn.readyOnce.Do(func() { close(conn.readyCh) })
}

// activate triggers the gateway-side handshake and blocks until the node confirms or the conn closes.
func (conn *standbyConn) activate() error {
	if conn.role != roleGateway {
		return errors.New("activate called on non-gateway conn")
	}
	conn.handoffOnce.Do(func() { close(conn.handoffCh) })
	select {
	case <-conn.readyCh:
		return nil
	case <-conn.doneCh:
		return ErrConnClosed
	}
}

func (conn *standbyConn) SetReadDeadline(t time.Time) error {
	if dl, ok := conn.Conn.(deadliner); ok {
		return dl.SetReadDeadline(t)
	}
	return nil
}

func (conn *standbyConn) SetWriteDeadline(t time.Time) error {
	if dl, ok := conn.Conn.(deadliner); ok {
		return dl.SetWriteDeadline(t)
	}
	return nil
}

// Close closes the underlying connection exactly once.
func (conn *standbyConn) Close() error {
	if conn.closed.Swap(true) {
		return nil
	}
	err := conn.Conn.Close()
	close(conn.doneCh)
	return err
}

// runKeepAlive runs the protocol loop. Context cancellation closes the connection.
func (conn *standbyConn) runKeepAlive(ctx context.Context) {
	if conn.role == roleClient {
		conn.runNodeSide(ctx)
	} else {
		conn.runGateway(ctx)
	}

	if !conn.relaying.Load() {
		conn.Close()
	}
}

func (conn *standbyConn) runNodeSide(ctx context.Context) {
	ch := channel.New(conn)
	if err := ch.Send(&Ping{}); err != nil {
		return
	}
	conn.SetReadDeadline(time.Now().Add(pingTimeout))
	err := ch.Switch(
		func(p *Ping) error {
			if !p.Pong {
				return errors.New("unexpected ping")
			}
			go conn.schedulePing(ch)
			return nil
		},
		func(s *Handoff) error {
			if s.Confirm {
				return errors.New("unexpected ack")
			}
			conn.claimed.Store(true)
			if err := ch.Send(&Handoff{Confirm: true}); err != nil {
				return err
			}
			conn.relaying.Store(true)
			conn.SetReadDeadline(time.Time{})
			conn.SetWriteDeadline(time.Time{})
			conn.markReady()
			conn.log.Log("relay started (client side) remote=%v", conn.RemoteEndpoint())
			return channel.ErrBreak
		},
		channel.WithContext(ctx),
	)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			conn.log.Log("conn timed out (read deadline) remote=%v", conn.RemoteEndpoint())
		}
		conn.Close()
	}
}

func (conn *standbyConn) runGateway(ctx context.Context) {
	ch := channel.New(conn)
	conn.SetReadDeadline(time.Now().Add(silenceTimeout))

	// Cancel the ping loop as soon as a handoff is requested, without waiting
	// for the next ping from the node (which could be up to pingInterval away).
	loopCtx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-conn.handoffCh:
		case <-loopCtx.Done():
			return
		}
		cancel()
	}()

	ch.Switch(
		func(p *Ping) error {
			if p.Pong {
				return errors.New("unexpected pong")
			}
			conn.SetReadDeadline(time.Now().Add(pingTimeout))
			return ch.Send(&Ping{Pong: true})
		},
		channel.WithContext(loopCtx),
	)
	cancel()

	// Proceed with handoff only if it was requested (not a timeout or parent cancel).
	select {
	case <-conn.handoffCh:
	default:
		return
	}

	if err := ch.Send(&Handoff{}); err != nil {
		return
	}

	ch.Switch(
		func(s *Handoff) error {
			if !s.Confirm {
				return errors.New("unexpected signal")
			}
			conn.relaying.Store(true)
			conn.SetReadDeadline(time.Time{})
			conn.SetWriteDeadline(time.Time{})
			conn.markReady()
			return channel.ErrBreak
		},
		channel.WithContext(ctx),
	)
}

func (conn *standbyConn) schedulePing(ch *channel.Channel) {
	select {
	case <-conn.readyCh:
		return
	case <-conn.doneCh:
		return
	case <-time.After(pingInterval):
	}
	if err := ch.Send(&Ping{}); err != nil {
		conn.Close()
		return
	}

	conn.SetReadDeadline(time.Now().Add(pingTimeout))
}
