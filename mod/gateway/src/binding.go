package gateway

import (
	"sync/atomic"

	"github.com/cryptopunkscc/astrald/astral"
	libastrald "github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	gatewayClient "github.com/cryptopunkscc/astrald/mod/gateway/client"
)

func (mod *Module) bindToGateway(ctx *astral.Context, gatewayID *astral.Identity, visibility gateway.Visibility) {
	client := gatewayClient.New(gatewayID, libastrald.Default())

	socket, err := client.Bind(ctx, visibility)
	if err != nil {
		mod.log.Error("bind to %v: %v", gatewayID, err)
		return
	}

	newGatewayBinding(mod, socket).Run(ctx)
}

type gatewayBinding struct {
	*Module
	socket *gateway.Socket
	count  atomic.Int32
}

func newGatewayBinding(module *Module, socket *gateway.Socket) *gatewayBinding {
	return &gatewayBinding{Module: module, socket: socket}
}

func (b *gatewayBinding) Run(ctx *astral.Context) {
	for range b.config.InitConns {
		b.spawn(ctx)
	}

	<-ctx.Done()
}

func (b *gatewayBinding) spawn(ctx *astral.Context) {
	if b.count.Add(1) > b.config.MaxConns {
		b.count.Add(-1)
		b.log.Error("max connections reached (%v), cannot spawn new slot", b.config.MaxConns)
		return
	}

	go func() {
		defer b.count.Add(-1)
		b.hold(ctx)
	}()
}

func (b *gatewayBinding) hold(ctx *astral.Context) {
	conn, err := b.Exonet.Dial(ctx, b.socket.Endpoint)
	if err != nil {
		return
	}

	// Authenticate with the gateway
	if _, err = b.socket.Nonce.WriteTo(conn); err != nil {
		b.log.Errorv(1, "nonce write to %v failed: %v", conn.RemoteEndpoint(), err)
		return
	}

	// Wrap conn: on first incoming byte, spawn a replacement slot; pass all bytes through untouched
	slot := &triggerConn{
		Conn:    conn,
		onFirst: func() { b.spawn(ctx) },
	}

	if err = b.Nodes.EstablishInboundLink(ctx, slot); err != nil {
		b.log.Errorv(1, "inbound link from %v failed: %v", conn.RemoteEndpoint(), err)
	}
}

// triggerConn wraps an exonet.Conn and calls onFirst exactly once on the first incoming bytes.
type triggerConn struct {
	exonet.Conn
	triggered atomic.Bool
	onFirst   func()
}

func (c *triggerConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if n > 0 && c.triggered.CompareAndSwap(false, true) {
		c.onFirst()
	}
	return n, err
}
