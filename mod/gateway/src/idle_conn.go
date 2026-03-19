package gateway

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
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

// note: maybe can be part of exonet
type deadliner interface {
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

type idleConn struct {
	exonet.Conn
	role         connRole
	log          *log.Logger
	withIdentity *astral.Identity // identity of peer we are connected to

	closed atomic.Bool

	handoffCh   chan struct{} // gateway only
	handoffOnce sync.Once

	readyCh   chan struct{}
	doneCh    chan struct{}
	readyOnce sync.Once
	doneOnce  sync.Once
}

func newIdleConn(conn exonet.Conn, role connRole, identity *astral.Identity, l *log.Logger) *idleConn {
	c := &idleConn{
		Conn:         conn,
		role:         role,
		log:          l,
		withIdentity: identity,
		readyCh:      make(chan struct{}),
		doneCh:       make(chan struct{}),
	}
	if role == roleGateway {
		c.handoffCh = make(chan struct{})
	}
	l.Logv(2, "idle conn created with %v remote %v", identity, conn.RemoteEndpoint())
	return c
}

func (c *idleConn) Ready() <-chan struct{} { return c.readyCh }
func (c *idleConn) Done() <-chan struct{}  { return c.doneCh }

func (c *idleConn) markReady() {
	c.readyOnce.Do(func() { close(c.readyCh) })
}

func (c *idleConn) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	err := c.Conn.Close()
	c.doneOnce.Do(func() { close(c.doneCh) })
	return err
}

func (c *idleConn) activate(ctx context.Context) error {
	if c.role != roleGateway {
		return errors.New("activate called on non-gateway conn")
	}
	c.handoffOnce.Do(func() { close(c.handoffCh) })

	select {
	case <-c.readyCh:
		return nil
	case <-c.doneCh:
		return ErrConnClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *idleConn) setReadDeadline(t time.Time) {
	if dl, ok := c.Conn.(deadliner); ok {
		dl.SetReadDeadline(t)
	}
}

func (c *idleConn) setWriteDeadline(t time.Time) {
	if dl, ok := c.Conn.(deadliner); ok {
		dl.SetWriteDeadline(t)
	}
}

func (c *idleConn) eventLoop(ctx context.Context) {
	ch := channel.New(c.Conn)

	lastActivity := time.Now()
	lastPing := time.Time{}
	var handoffDone bool
	var handoffSuccess bool

	defer func() {
		if !handoffSuccess {
			c.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		now := time.Now()

		if c.role == roleGateway && !handoffDone {
			select {
			case <-c.handoffCh:
				c.setWriteDeadline(now.Add(writeTimeout))
				if err := ch.Send(&Handoff{}); err != nil {
					return
				}
				handoffDone = true
				lastActivity = now
				lastPing = now
			default:
			}
		}

		if now.Sub(lastActivity) >= pingInterval && now.Sub(lastPing) >= pingInterval {
			c.setWriteDeadline(now.Add(writeTimeout))
			if err := ch.Send(&Ping{}); err != nil {
				return
			}
			lastPing = now
		}

		readWait := pingTimeout
		if c.role == roleGateway && !handoffDone {
			readWait = handoffPollInterval
		}

		c.setReadDeadline(now.Add(readWait))

		obj, err := ch.Receive()
		if err != nil {
			if isTimeout(err) {
				if time.Since(lastActivity) >= pingTimeout {
					c.log.Logv(2, "closing idle conn with %v idle for %v", c.withIdentity, time.Since(lastActivity).Round(time.Second).String())
					return
				}
				continue
			}

			return
		}

		lastActivity = time.Now()

		switch m := obj.(type) {
		case *Ping:
			if !m.Pong {
				c.setWriteDeadline(time.Now().Add(writeTimeout))
				if err := ch.Send(&Ping{Pong: true}); err != nil {
					return
				}
			}

		case *Handoff:
			c.setWriteDeadline(time.Now().Add(writeTimeout))
			if err := ch.Send(&HandoffAck{}); err != nil {
				return
			}
			handoffSuccess = true
			c.setReadDeadline(time.Time{})
			c.setWriteDeadline(time.Time{})
			c.markReady()
			return

		case *HandoffAck:
			handoffSuccess = true
			c.setReadDeadline(time.Time{})
			c.setWriteDeadline(time.Time{})
			c.markReady()
			return

		default:
			return
		}
	}
}

func isTimeout(err error) bool {
	ne, ok := err.(net.Error)
	return ok && ne.Timeout()
}
