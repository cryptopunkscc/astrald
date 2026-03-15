package gateway

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/sig"
)

const (
	socketPoolTargetIdle = 2
	socketPoolMaxFails   = 3
	socketPingInterval   = 2 * time.Second
	socketPingTimeout    = 3 * time.Second
)

// SocketPool maintains socketPoolTargetIdle idle socket connections to a gateway.
type SocketPool struct {
	*Module
	ctx       *astral.Context
	socket    gateway.Socket
	gatewayID *astral.Identity

	conns sig.Set[*bindingConn]
	wake  chan struct{}
}

func (mod *Module) newSocketPool(ctx *astral.Context, gatewayID *astral.Identity, socket gateway.Socket) *SocketPool {
	return &SocketPool{
		ctx:       ctx,
		Module:    mod,
		socket:    socket,
		gatewayID: gatewayID,
		wake:      make(chan struct{}, 1),
	}
}

func (p *SocketPool) Run() error {
	retry, _ := sig.NewRetry(time.Second, 30*time.Second, 2)
	p.notify()

	for {
		select {
		case <-p.ctx.Done():
			return p.ctx.Err()
		case <-p.wake:
			for p.idleCount() < socketPoolTargetIdle {
				conn, err := p.acquireConn()
				if err != nil {
					select {
					case <-p.ctx.Done():
						return p.ctx.Err()
					case count := <-retry.Retry():
						if count >= socketPoolMaxFails {
							return gateway.ErrSocketUnreachable
						}
					}
					continue
				}
				retry.Reset()
				p.startIdleSocket(conn)
			}
		}
	}
}

func (p *SocketPool) acquireConn() (exonet.Conn, error) {
	p.log.Logv(2, "acquiring socket connection to %v through %v", p.gatewayID, p.socket.Endpoint)
	conn, err := p.Exonet.Dial(p.ctx, p.socket.Endpoint)
	if err != nil {
		return nil, err
	}

	if _, err := p.socket.Nonce.WriteTo(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

// idleCount returns the number of non-active conns in the pool.
func (p *SocketPool) idleCount() int {
	return len(p.conns.Select(func(a *bindingConn) bool {
		return !a.active.Load()
	}))
}

func (p *SocketPool) startIdleSocket(conn exonet.Conn) {
	bc := newBinderConn(conn)
	bc.onClose = func() {
		p.conns.Remove(bc)
		p.notify()
	}

	p.conns.Add(bc)
	go bc.keepalive(p.ctx.Done(), func() error {
		return p.Nodes.EstablishInboundLink(p.ctx, &gwConn{
			ReadWriteCloser: bc,
			local:           gateway.NewEndpoint(p.node.Identity(), p.node.Identity()),
			remote:          gateway.NewEndpoint(p.gatewayID, p.node.Identity()),
		})
	})
}

func (p *SocketPool) notify() {
	select {
	case p.wake <- struct{}{}:
	default:
	}
}
