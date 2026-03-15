package gateway

import (
	"sync"
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

	mu   sync.Mutex
	idle int

	wake chan struct{}
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
			for {
				p.mu.Lock()
				if p.idle >= socketPoolTargetIdle {
					p.mu.Unlock()
					break
				}
				p.mu.Unlock()

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

func (p *SocketPool) addIdle() {
	p.mu.Lock()
	p.idle++
	p.mu.Unlock()
}

func (p *SocketPool) onConnTaken() {
	p.mu.Lock()
	p.idle--
	p.mu.Unlock()
	p.notify()
}

func (p *SocketPool) onConnClosed(idle bool) {
	if idle {
		p.mu.Lock()
		p.idle--
		p.mu.Unlock()
	}
	p.notify()
}

func (p *SocketPool) startIdleSocket(conn exonet.Conn) {
	bc := newBinderConn(conn, nil)
	gc := &gwConn{
		bindingConn: bc,
		local:       gateway.NewEndpoint(p.node.Identity(), p.node.Identity()),
		remote:      gateway.NewEndpoint(p.gatewayID, p.node.Identity()),
	}
	bc.onClose = func() { p.onConnClosed(!bc.active.Load()) }
	p.addIdle()
	go bc.keepalive(p.ctx.Done(), p.onConnTaken, func() error {
		return p.Nodes.EstablishInboundLink(p.ctx, gc)
	})
}

func (p *SocketPool) notify() {
	select {
	case p.wake <- struct{}{}:
	default:
	}
}
