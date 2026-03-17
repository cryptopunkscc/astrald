package gateway

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

var ErrConnClosed = errors.New("conn closed")

type connRole uint8

const (
	roleClient connRole = iota
	roleGateway
)

// standbyConn is a unified idle socket connection for both node and gateway sides.
type standbyConn struct {
	exonet.Conn
	role connRole

	closed   atomic.Bool
	claimed  atomic.Bool
	relaying atomic.Bool

	handoffCh   chan struct{}
	handoffOnce sync.Once
	readyCh     chan struct{}
	doneCh      chan struct{}
	readyOnce   sync.Once
}

func newGatewayConn(conn exonet.Conn, role connRole) *standbyConn {
	bc := &standbyConn{
		Conn:    conn,
		role:    role,
		readyCh: make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
	if role == roleGateway {
		bc.handoffCh = make(chan struct{})
	}
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
	defer func() {
		if !conn.relaying.Load() {
			conn.Close()
		}
	}()

	if conn.role == roleClient {
		conn.runNodeSide(ctx)
		return
	}
	conn.runGateway(ctx)
}

func (conn *standbyConn) runNodeSide(ctx context.Context) {
	ch := channel.New(conn)
	if err := ch.Send(&Ping{}); err != nil {
		return
	}
	conn.SetReadDeadline(time.Now().Add(pingTimeout))
	ch.Switch(
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
			conn.markReady()
			return channel.ErrBreak
		},
		channel.WithContext(ctx),
	)
}

func (conn *standbyConn) runGateway(ctx context.Context) {
	ch := channel.New(conn)
	conn.SetReadDeadline(time.Now().Add(silenceTimeout))
	ch.Switch(
		func(p *Ping) error {
			if p.Pong {
				return errors.New("unexpected pong")
			}
			conn.SetReadDeadline(time.Now().Add(silenceTimeout))
			select {
			case <-conn.handoffCh:
				return ch.Send(&Handoff{})
			default:
				return ch.Send(&Ping{Pong: true})
			}
		},
		func(s *Handoff) error {
			if !s.Confirm {
				return errors.New("unexpected signal")
			}
			conn.relaying.Store(true)
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
