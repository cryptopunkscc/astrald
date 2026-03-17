package gateway

import (
	"context"
	"errors"
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

type state uint8

const (
	idle        state = iota
	waitConfirm state = iota
)

type standbyConn struct {
	exonet.Conn
	role     connRole
	log      *log.Logger
	identity *astral.Identity

	closed atomic.Bool

	handoffCh   chan struct{} // gateway only
	handoffOnce sync.Once

	readyCh   chan struct{}
	doneCh    chan struct{}
	readyOnce sync.Once
	doneOnce  sync.Once
}

func newGatewayConn(conn exonet.Conn, role connRole, identity *astral.Identity, l *log.Logger) *standbyConn {
	c := &standbyConn{
		Conn:     conn,
		role:     role,
		log:      l,
		identity: identity,
		readyCh:  make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
	if role == roleGateway {
		c.handoffCh = make(chan struct{})
	}
	l.Logv(2, "standby conn created identity=%v remote=%v", identity, conn.RemoteEndpoint())
	return c
}

func (c *standbyConn) markReady() {
	c.readyOnce.Do(func() { close(c.readyCh) })
}

func (c *standbyConn) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	err := c.Conn.Close()
	c.doneOnce.Do(func() { close(c.doneCh) })
	return err
}

func (c *standbyConn) activate(ctx context.Context) error {
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

func (c *standbyConn) setReadDeadline(t time.Time) {
	if dl, ok := c.Conn.(deadliner); ok {
		dl.SetReadDeadline(t)
	}
}

func (c *standbyConn) setWriteDeadline(t time.Time) {
	if dl, ok := c.Conn.(deadliner); ok {
		dl.SetWriteDeadline(t)
	}
}

func (c *standbyConn) eventLoop(ctx context.Context) {
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
			c.log.Logv(1, "ctx done remote=%v", c.Conn.RemoteEndpoint())
			return
		default:
		}

		now := time.Now()

		// gateway activation trigger
		if c.role == roleGateway && !handoffDone {
			select {
			case <-c.handoffCh:
				c.log.Logv(1, "send Handoff remote=%v", c.Conn.RemoteEndpoint())
				c.setWriteDeadline(now.Add(writeTimeout))
				if err := ch.Send(&Handoff{}); err != nil {
					c.log.Errorv(1, "send handoff failed remote=%v err=%v", c.Conn.RemoteEndpoint(), err)
					return
				}
				handoffDone = true
				lastActivity = now
				lastPing = now
			default:
			}
		}

		// keepalive only when idle enough
		if now.Sub(lastActivity) >= pingInterval && now.Sub(lastPing) >= pingInterval {
			c.log.Logv(3, "send Ping remote=%v", c.Conn.RemoteEndpoint())
			c.setWriteDeadline(now.Add(writeTimeout))
			if err := ch.Send(&Ping{}); err != nil {
				c.log.Errorv(1, "send ping failed remote=%v err=%v", c.Conn.RemoteEndpoint(), err)
				return
			}
			lastPing = now
		}

		// adaptive read wait
		readWait := pingTimeout
		if c.role == roleGateway && !handoffDone {
			if readWait > time.Second {
				readWait = time.Second
			}
		}

		c.setReadDeadline(now.Add(readWait))

		obj, err := ch.Receive()
		if err != nil {
			if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
				c.log.Logv(2, "timeout idle=%v remote=%v", time.Since(lastActivity), c.Conn.RemoteEndpoint())
				if time.Since(lastActivity) >= pingTimeout {
					c.log.Errorv(1, "idle timeout close remote=%v", c.Conn.RemoteEndpoint())
					return
				}
				continue
			}
			c.log.Errorv(1, "receive error remote=%v err=%v", c.Conn.RemoteEndpoint(), err)
			return
		}

		c.log.Logv(2, "recv %T remote=%v", obj, c.Conn.RemoteEndpoint())
		lastActivity = time.Now()

		switch m := obj.(type) {

		case *Ping:
			if !m.Pong {
				c.log.Logv(3, "recv Ping → send Pong remote=%v", c.Conn.RemoteEndpoint())
				c.setWriteDeadline(time.Now().Add(writeTimeout))
				if err := ch.Send(&Ping{Pong: true}); err != nil {
					c.log.Errorv(1, "send pong failed remote=%v err=%v", c.Conn.RemoteEndpoint(), err)
					return
				}
			} else {
				c.log.Logv(3, "recv Pong remote=%v", c.Conn.RemoteEndpoint())
			}

		case *Handoff:
			if !m.Confirm {
				c.log.Logv(1, "recv Handoff → send Confirm remote=%v", c.Conn.RemoteEndpoint())
				c.setWriteDeadline(time.Now().Add(writeTimeout))
				if err := ch.Send(&Handoff{Confirm: true}); err != nil {
					c.log.Errorv(1, "send confirm failed remote=%v err=%v", c.Conn.RemoteEndpoint(), err)
					return
				}
			} else {
				c.log.Logv(1, "recv HandoffConfirm remote=%v", c.Conn.RemoteEndpoint())
			}
			c.markReady()
			return

		default:
			c.log.Errorv(1, "unexpected frame %T remote=%v", obj, c.Conn.RemoteEndpoint())
			return
		}
	}
}
