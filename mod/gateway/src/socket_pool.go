package gateway

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

const (
	socketPoolTargetIdle = 1
	socketPoolMaxFails   = 3
)

var ErrSocketUnreachable = errors.New("socket unreachable")

type SocketPool struct {
	ctx    *astral.Context
	socket *gateway.Socket
	exonet exonet.Module
	nodes  nodes.Module
	log    *log.Logger

	mu    sync.Mutex
	total int
	idle  int

	signal chan struct{}
}

func newSocketPool(ctx *astral.Context, mod *Module, socket *gateway.Socket) *SocketPool {
	return &SocketPool{
		ctx:    ctx,
		socket: socket,
		exonet: mod.Exonet,
		nodes:  mod.Nodes,
		log:    mod.log,
		signal: make(chan struct{}, 1),
	}
}

func (p *SocketPool) Run() error {
	retry, _ := sig.NewRetry(time.Second, 2*time.Minute, 2)
	failStreak := 0

	p.notify()

	for {
		select {
		case <-p.ctx.Done():
			return p.ctx.Err()
		case <-p.signal:
			for p.idleCount() < socketPoolTargetIdle {
				conn, err := p.acquireSocketConnection()
				if err != nil {
					failStreak++
					if failStreak >= socketPoolMaxFails {
						p.log.Log("gateway socket %v unreachable", p.socket.Endpoint)
						return ErrSocketUnreachable
					}
					select {
					case <-p.ctx.Done():
						return p.ctx.Err()
					case <-retry.Retry():
					}
					continue
				}

				failStreak = 0
				retry.Reset()
				p.attach(conn)
			}
		}
	}
}

func (p *SocketPool) acquireSocketConnection() (exonet.Conn, error) {
	conn, err := p.exonet.Dial(p.ctx, p.socket.Endpoint)
	if err != nil {
		return nil, err
	}
	if _, err := p.socket.Nonce.WriteTo(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func (p *SocketPool) addIdle() (total, idle int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.total++
	p.idle++
	return p.total, p.idle
}

func (p *SocketPool) connTaken() (total, idle int) {
	p.mu.Lock()
	p.idle--
	total, idle = p.total, p.idle
	p.mu.Unlock()
	p.notify()
	return
}

func (p *SocketPool) onClosedConn(wasUsed bool) (total, idle int) {
	p.mu.Lock()
	p.total--
	if !wasUsed {
		p.idle--
	}
	total, idle = p.total, p.idle
	p.mu.Unlock()
	if !wasUsed {
		p.notify()
	}
	return
}

func (p *SocketPool) idleCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.idle
}

func (p *SocketPool) attach(conn exonet.Conn) {
	total, idle := p.addIdle()
	p.log.Logv(1, "gateway conn up (total: %v, idle: %v)", total, idle)

	pc := &socketConn{Conn: conn}

	pc.onFirst = func() {
		total, idle := p.connTaken()
		p.log.Logv(1, "gateway conn taken (total: %v, idle: %v)", total, idle)
	}

	pc.onClose = func() {
		total, idle := p.onClosedConn(pc.used.Load())
		p.log.Logv(1, "gateway conn down (total: %v, idle: %v)", total, idle)
	}

	go func() {
		if err := p.nodes.EstablishInboundLink(p.ctx, pc); err != nil {
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
