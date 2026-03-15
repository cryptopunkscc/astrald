package gateway

import (
	"io"
	"sync"
	"sync/atomic"
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

type SocketPool struct {
	*Module
	ctx       *astral.Context
	socket    gateway.Socket
	gatewayID *astral.Identity

	mu    sync.Mutex
	total int
	idle  int

	signal chan struct{}
}

func (mod *Module) newSocketPool(ctx *astral.Context, gatewayID *astral.Identity, socket gateway.Socket) *SocketPool {
	return &SocketPool{
		ctx:       ctx,
		Module:    mod,
		socket:    socket,
		gatewayID: gatewayID,
		signal:    make(chan struct{}, 1),
	}
}

func (p *SocketPool) Run() error {
	retry, _ := sig.NewRetry(time.Second, 30*time.Second, 2)
	p.notify()

	for {
		select {
		case <-p.ctx.Done():
			return p.ctx.Err()
		case <-p.signal:
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
				p.handoff(conn)
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

func (p *SocketPool) idleCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.idle
}

func (p *SocketPool) addIdle() {
	p.mu.Lock()
	p.idle++
	p.total++
	p.mu.Unlock()
}

func (p *SocketPool) onConnTaken() {
	p.mu.Lock()
	p.idle--
	p.mu.Unlock()
	p.notify()
}

func (p *SocketPool) onConnClosed(wasIdle bool) {
	p.mu.Lock()
	p.total--
	if wasIdle {
		p.idle--
	}
	p.mu.Unlock()
	p.notify()
}

func (p *SocketPool) handoff(conn exonet.Conn) {
	pc := &socketConn{
		Conn:           conn,
		localEndpoint:  gateway.NewEndpoint(p.node.Identity(), p.node.Identity()),
		remoteEndpoint: gateway.NewEndpoint(p.gatewayID, p.node.Identity()),
	}
	p.registerConn(pc)
	go p.runIdleConn(conn, pc)
}

// registerConn binds pool lifecycle callbacks to pc and marks it as idle.
func (p *SocketPool) registerConn(pc *socketConn) {
	pc.onClose = func() { p.onConnClosed(!pc.used.Load()) }
	p.addIdle()
}

// runIdleConn is the binder-side ping loop for an idle socket connection.
// It sends BytePing and waits for BytePong or ByteSignalGo.
// On ByteSignalGo the conn transitions from idle to taken before writing ByteSignalReady.
func (p *SocketPool) runIdleConn(conn exonet.Conn, pc *socketConn) {
	for {
		if err := gateway.WritePing(conn); err != nil {
			pc.Close()
			return
		}
		withReadDeadline(conn, socketPingTimeout)
		var b [1]byte
		if _, err := io.ReadFull(conn, b[:]); err != nil {
			pc.Close()
			return
		}
		clearDeadlines(conn)
		switch b[0] {
		case gateway.BytePong:
			select {
			case <-time.After(socketPingInterval):
			case <-p.ctx.Done():
				pc.Close()
				return
			}
		case gateway.ByteSignalGo:
			// Mark taken before notifying pool: if Close fires during write,
			// onClose sees used=true → wasIdle=false → only total is decremented.
			pc.used.Store(true)
			p.onConnTaken()
			if err := gateway.WriteSignalReady(pc); err != nil {
				pc.Close()
				return
			}
			if err := p.Nodes.EstablishInboundLink(p.ctx, pc); err != nil {
				pc.Close()
			}
			return
		default:
			pc.Close()
			return
		}
	}
}

func (p *SocketPool) notify() {
	select {
	case p.signal <- struct{}{}:
	default:
	}
}

// socketConn wraps an exonet.Conn with fixed endpoints and a one-shot close callback.
// used tracks whether the conn has been taken from the idle pool (for wasIdle accounting).
type socketConn struct {
	exonet.Conn

	localEndpoint  exonet.Endpoint
	remoteEndpoint exonet.Endpoint
	onClose        func()

	closed atomic.Bool
	used   atomic.Bool
}

func (c *socketConn) LocalEndpoint() exonet.Endpoint  { return c.localEndpoint }
func (c *socketConn) RemoteEndpoint() exonet.Endpoint { return c.remoteEndpoint }

func (c *socketConn) Close() error {
	err := c.Conn.Close()
	if !c.closed.Swap(true) && c.onClose != nil {
		c.onClose()
	}
	return err
}
