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
	l.Logv(2, "standby conn created role=%v identity=%v remote=%v", role, identity, conn.RemoteEndpoint())
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

	var s = idle
	var nextPing = time.Now().Add(pingInterval)

	defer c.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Gateway: check for a pending handoff request before each send/receive.
		if c.role == roleGateway && s == idle {
			select {
			case <-c.handoffCh:
				c.setWriteDeadline(time.Now().Add(writeTimeout))
				if err := ch.Send(&Handoff{}); err != nil {
					return
				}
				s = waitConfirm
			default:
			}
		}

		// Send ping only when idle and cooldown has elapsed.
		if s == idle && time.Now().After(nextPing) {
			c.setWriteDeadline(time.Now().Add(writeTimeout))
			if err := ch.Send(&Ping{}); err != nil {
				return
			}
			nextPing = time.Now().Add(pingInterval)
		}

		// Wait for the next message.
		c.setReadDeadline(time.Now().Add(pingTimeout))

		obj, err := ch.Receive()
		if err != nil {
			return
		}

		// Any received message proves peer is alive; push back the ping deadline.
		nextPing = time.Now().Add(pingInterval)

		switch m := obj.(type) {

		case *Ping:
			if !m.Pong {
				c.setWriteDeadline(time.Now().Add(writeTimeout))
				if err := ch.Send(&Ping{Pong: true}); err != nil {
					return
				}
				continue
			}
			continue

		case *Handoff:
			switch s {
			case idle:
				if m.Confirm {
					return
				}
				c.setWriteDeadline(time.Now().Add(writeTimeout))
				if err := ch.Send(&Handoff{Confirm: true}); err != nil {
					return
				}
				c.markReady()
				return

			case waitConfirm:
				if !m.Confirm {
					return
				}
				c.markReady()
				return

			default:
				return
			}

		default:
			return
		}
	}
}
