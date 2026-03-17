package gateway

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/sig"
)

// ConnPool maintains minIdleConns idle socket connections to a gateway.
type ConnPool struct {
	*Module
	ctx       *astral.Context
	socket    gateway.Socket
	gatewayID *astral.Identity

	conns    sig.Set[*standbyConn]
	refillCh chan struct{}
}

func (mod *Module) newConnPool(ctx *astral.Context, gatewayID *astral.Identity, socket gateway.Socket) *ConnPool {
	return &ConnPool{
		ctx:       ctx,
		Module:    mod,
		socket:    socket,
		gatewayID: gatewayID,
		refillCh:  make(chan struct{}, 1),
	}
}

func (p *ConnPool) Run() error {
	retry, err := sig.NewRetry(time.Second, 30*time.Second, 2)
	if err != nil {
		return err
	}
	p.notify()

	for {
		select {
		case <-p.ctx.Done():
			return p.ctx.Err()
		case <-p.refillCh:
			for p.idleCount() < minIdleConns {
				conn, err := p.dialSocket()
				if err != nil {
					select {
					case <-p.ctx.Done():
						return p.ctx.Err()
					case count := <-retry.Retry():
						if count >= maxDialFails {
							return gateway.ErrSocketUnreachable
						}
					}
					continue
				}
				retry.Reset()
				p.addIdleConn(conn)
			}
		}
	}
}

func (p *ConnPool) dialSocket() (exonet.Conn, error) {
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

// idleCount returns the number of non-claimed conns in the pool.
func (p *ConnPool) idleCount() int {
	return len(p.conns.Select(func(a *standbyConn) bool {
		return !a.claimed.Load()
	}))
}

func (p *ConnPool) addIdleConn(conn exonet.Conn) {
	bc := newGatewayConn(conn, roleClient)
	p.conns.Add(bc)

	go func() {
		<-bc.doneCh
		p.conns.Remove(bc)
		p.notify()
	}()

	go func() {
		select {
		case <-bc.readyCh:
			if err := p.Nodes.EstablishInboundLink(p.ctx, &gatewayConn{
				ReadWriteCloser: bc,
				local:           gateway.NewEndpoint(p.node.Identity(), p.node.Identity()),
				remote:          gateway.NewEndpoint(p.gatewayID, p.node.Identity()),
			}); err != nil {
				bc.Close()
			}
		case <-bc.doneCh:
		}
	}()

	go bc.runKeepAlive(p.ctx)
}

func (p *ConnPool) notify() {
	select {
	case p.refillCh <- struct{}{}:
	default:
	}
}
