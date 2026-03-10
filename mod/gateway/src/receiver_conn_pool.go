package gateway

import (
	"sync/atomic"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

type receiverConnPool struct {
	*Module
	socket *gateway.Socket
	count  atomic.Int32
}

func newReceiverConnPool(module *Module, socket *gateway.Socket) *receiverConnPool {
	return &receiverConnPool{Module: module, socket: socket}
}

func (pool *receiverConnPool) Run(ctx *astral.Context) {
	for range pool.config.InitConns {
		pool.spawn(ctx)
	}

	<-ctx.Done()
}

func (pool *receiverConnPool) spawn(ctx *astral.Context) {
	if pool.count.Add(1) > pool.config.MaxConns {
		pool.count.Add(-1)
		pool.log.Error("max connections reached (%v), cannot spawn new slot", pool.config.MaxConns)
		return
	}

	go func() {
		defer pool.count.Add(-1)
		pool.hold(ctx)
	}()
}

func (pool *receiverConnPool) hold(ctx *astral.Context) {
	conn, err := pool.Exonet.Dial(ctx, pool.socket.Endpoint)
	if err != nil {
		return
	}

	// Authenticate with the gateway
	if _, err = pool.socket.Nonce.WriteTo(conn); err != nil {
		pool.log.Errorv(1, "nonce write to %v failed: %v", conn.RemoteEndpoint(), err)
		return
	}

	// Wrap conn: on first incoming byte, spawn a replacement slot; pass all bytes through untouched
	slot := &slotConn{
		Conn:    conn,
		onFirst: func() { pool.spawn(ctx) },
	}

	if err = pool.Nodes.EstablishInboundLink(ctx, slot); err != nil {
		pool.log.Errorv(1, "inbound link from %v failed: %v", conn.RemoteEndpoint(), err)
	}
}

// slotConn wraps an exonet.Conn and calls onFirst exactly once on the first incoming bytes.
type slotConn struct {
	exonet.Conn
	triggered atomic.Bool
	onFirst   func()
}

func (c *slotConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if n > 0 && c.triggered.CompareAndSwap(false, true) {
		c.onFirst()
	}
	return n, err
}
