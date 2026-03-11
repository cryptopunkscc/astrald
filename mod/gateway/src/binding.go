package gateway

import (
	"fmt"
	"sync/atomic"

	"github.com/cryptopunkscc/astrald/astral"
	libastrald "github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	gatewayClient "github.com/cryptopunkscc/astrald/mod/gateway/client"
)

func (mod *Module) bindToGateway(ctx *astral.Context, gatewayID *astral.Identity, visibility gateway.Visibility) (err error) {
	mod.log.Logv(1, "binding to gateway %v", gatewayID)
	client := gatewayClient.New(gatewayID, libastrald.Default())

	socket, err := client.Bind(ctx.IncludeZone(astral.ZoneNetwork), visibility)
	if err != nil {
		fmt.Println("BIND FAILED: ", err)
		return err
	}

	fmt.Println("GOT SOCKET: ", socket)
	// todo: if we lose connection to gateway (e.g reboot) we should try to rebind

	go newGatewayBinding(mod, socket).Run(ctx)

	return nil
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

	fmt.Println("DIALED GW")
	// Authenticate with the gateway
	if _, err = b.socket.Nonce.WriteTo(conn); err != nil {
		b.log.Errorv(1, "nonce write to %v failed: %v", conn.RemoteEndpoint(), err)
		return
	}

	fmt.Println("WROTE NONCE")

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
