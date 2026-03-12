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
	socketPoolTargetIdle = 1
	socketPoolMaxFails   = 3
)

type SocketPool struct {
	*Module
	ctx    *astral.Context
	socket *gateway.Socket
	log    *log.Logger

	mu    sync.Mutex
	total int
	idle  int

	signal chan struct{}
}

func (mod *Module) newSocketPool(ctx *astral.Context, socket *gateway.Socket) *SocketPool {
	return &SocketPool{
		ctx:    ctx,
		Module: mod,
		socket: socket,
		log:    mod.log,
		signal: make(chan struct{}, 1),
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
							p.log.Log("gateway socket %v unreachable", p.socket.Endpoint)
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
	total, idle := p.total, p.idle
	p.mu.Unlock()
	p.notify()
	p.log.Logv(1, "gateway conn added (total: %v, idle: %v)", total, idle)
}

func (p *SocketPool) onConnTaken() {
	p.mu.Lock()
	p.idle--
	total, idle := p.total, p.idle
	p.mu.Unlock()
	p.notify()
	p.log.Logv(1, "gateway conn taken (total: %v, idle: %v)", total, idle)
}

func (p *SocketPool) onConnClosed() {
	p.mu.Lock()
	p.total--
	total, idle := p.total, p.idle
	p.mu.Unlock()
	p.notify()
	p.log.Logv(1, "gateway conn down (total: %v, idle: %v)", total, idle)
}

func (p *SocketPool) handoff(conn exonet.Conn) {
	pc := &socketConn{Conn: conn}
	pc.onFirst = p.onConnTaken
	pc.onClose = p.onConnClosed
	p.addIdle()

	go func() {
		if err := p.Nodes.EstablishInboundLink(p.ctx, pc); err != nil {
			p.log.Logv(1, "inbound link from %v: %v", conn.RemoteEndpoint(), err)
		}
	}()
}

func (p *SocketPool) notify() {
	select {
	case p.signal <- struct{}{}:
	default:
	}
}

type socketConn struct {
	exonet.Conn

	onFirst func()
	onClose func()

	used atomic.Bool
}

func (c *socketConn) Read(b []byte) (int, error) {
	if !c.used.Swap(true) && c.onFirst != nil {
		c.onFirst()
	}
	return c.Conn.Read(b)
}

func (c *socketConn) Write(b []byte) (int, error) {
	return c.Conn.Write(b)
}

func (c *socketConn) Close() error {
	err := c.Conn.Close()
	if c.onClose != nil {
		c.onClose()
	}
	return err
}
