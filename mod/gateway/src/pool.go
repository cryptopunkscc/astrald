package gateway

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/sig"
)

const (
	socketPoolTargetIdle = 2
	socketPoolMaxFails   = 3
)

type SocketPool struct {
	*Module
	ctx       *astral.Context
	socket    *gateway.Socket
	gatewayID *astral.Identity
	log       *log.Logger

	mu    sync.Mutex
	total int
	idle  int

	signal chan struct{}
}

func (mod *Module) newSocketPool(ctx *astral.Context, gatewayID *astral.Identity, socket *gateway.Socket) *SocketPool {
	return &SocketPool{
		ctx:       ctx,
		Module:    mod,
		socket:    socket,
		gatewayID: gatewayID,
		log:       mod.log,
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
			for toAdd := socketPoolTargetIdle - p.idleCount(); toAdd > 0; toAdd-- {
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
					toAdd++
					continue
				}

				retry.Reset()
				p.handoff(conn)
			}
		}
	}
}

func (p *SocketPool) acquireConn() (exonet.Conn, error) {
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

	// when first write is done it means we started responding to link negotiation
	pc.onFirstWrite = p.onConnTaken
	pc.onClose = func() { p.onConnClosed(!pc.used.Load()) }
	p.addIdle()

	go func() {
		err := p.Nodes.EstablishInboundLink(p.ctx, pc)
		if err != nil {
			p.log.Logv(1, "inbound link from %v: %v", conn.RemoteEndpoint(), err)
			return
		}
	}()
}

func (p *SocketPool) notify() {
	select {
	case p.signal <- struct{}{}:
	default:
	}
}

// socketConn is considered a connection only after the first write is done.
type socketConn struct {
	exonet.Conn

	localEndpoint  exonet.Endpoint
	remoteEndpoint exonet.Endpoint
	onFirstWrite   func()
	onClose        func()

	closed atomic.Bool
	used   atomic.Bool
}

func (c *socketConn) LocalEndpoint() exonet.Endpoint  { return c.localEndpoint }
func (c *socketConn) RemoteEndpoint() exonet.Endpoint { return c.remoteEndpoint }

func (c *socketConn) Write(b []byte) (int, error) {
	if !c.used.Swap(true) && c.onFirstWrite != nil {
		c.onFirstWrite()
	}

	return c.Conn.Write(b)
}

func (c *socketConn) Close() error {
	err := c.Conn.Close()

	if !c.closed.Swap(true) {
		if c.onClose != nil {
			c.onClose()
		}
	}

	return err
}
