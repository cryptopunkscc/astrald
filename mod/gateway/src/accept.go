package gateway

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

func (mod *Module) startServers(ctx *astral.Context) {
	for _, addr := range mod.config.Gateway.Listen {
		parts := strings.SplitN(addr, ":", 2)
		if len(parts) != 2 {
			mod.log.Error("invalid listen address: %v", addr)
			continue
		}
		network, address := parts[0], parts[1]
		endpoint, err := mod.Exonet.Parse(network, address)
		if err != nil {
			mod.log.Error("parse listen address %v: %v", addr, err)
			continue
		}

		switch network {
		case "tcp":
			tcpEndpoint, ok := endpoint.(*tcp.Endpoint)
			if !ok {
				mod.log.Error("invalid listen address: %v", addr)
				continue
			}

			mod.log.Logv(1, "start listening on %v", tcpEndpoint)
			if err := mod.TCP.CreateEphemeralListener(ctx, tcpEndpoint.Port, mod.acceptSocketConn); err != nil {
				mod.log.Error("create ephemeral listener on %v: %v", addr, err)
				continue
			}

			mod.listenEndpoints.Set("tcp", tcpEndpoint)
		default:
			mod.log.Error("unsupported gateway socket network: %v", network)
		}
	}
}

// acceptSocketConn accepts connection on the socket that gateway told client to connect to.
func (mod *Module) acceptSocketConn(_ context.Context, conn exonet.Conn) (stopListener bool, err error) {
	mod.log.Logv(2, "accepting socket connection from %v", conn.RemoteEndpoint())

	var nonce astral.Nonce
	if _, err := nonce.ReadFrom(conn); err != nil {
		mod.log.Errorv(1, "read nonce from %v: %v", conn.RemoteEndpoint(), err)
		conn.Close()
		return stopListener, nil
	}

	if b, ok := mod.binderByNonce(nonce); ok {
		mod.log.Infov(2, "added idle conn to binder %v", b.Identity)
		bc := b.addConn(conn)
		go mod.keepalive(bc)
		return stopListener, nil
	}

	c, ok := mod.connectorByNonce(nonce)
	if !ok {
		mod.log.Errorv(1, "unknown nonce %v from %v", nonce, conn.RemoteEndpoint())
		conn.Close()
		return stopListener, nil
	}

	mod.connectors.Remove(c)

	reserved := c.takeReserved()
	if reserved == nil {
		conn.Close()
		return stopListener, fmt.Errorf("no reserved conn for %v", c.Target)
	}

	targetBinder, ok := mod.binderByIdentity(c.Target)
	if !ok {
		reserved.Close()
		conn.Close()
		return stopListener, nil
	}

	alive := mod.probeBinderConn(targetBinder, reserved)
	if alive == nil {
		mod.log.Errorv(1, "no alive conn for %v", c.Target)
		conn.Close()
		return stopListener, nil
	}

	cc := &connectorConn{
		Conn:    conn,
		network: conn.RemoteEndpoint().Network(),
		pipedTo: alive,
	}

	targetBinder.markPiped(alive, cc)
	mod.log.Infov(2, "pipe from %v to %v created", c.Identity, c.Target)
	go pipe(alive, cc)
	return stopListener, nil
}

const (
	socketPingInterval     = 2 * time.Second
	socketPingTimeout      = 3 * time.Second
	socketDeadTimeout      = 10 * time.Second
	socketProbeMaxAttempts = 3
	socketProbeTimeout     = 5 * time.Second
)

// keepalive runs on the gateway side for each idle binder conn.
// It reads binder pings and responds with pong. When a connector arrives
// via goCh it sends ByteSignalGo and waits for ByteSignalReady.
func (mod *Module) keepalive(bc *binderConn) {
	defer close(bc.done)
	piped := false
	defer func() {
		if !piped {
			bc.Close()
		}
	}()

	for {
		if d, ok := bc.Conn.(deadliner); ok {
			d.SetReadDeadline(time.Now().Add(socketDeadTimeout))
		}
		var b [1]byte
		if _, err := io.ReadFull(bc.Conn, b[:]); err != nil {
			return
		}
		if d, ok := bc.Conn.(deadliner); ok {
			d.SetReadDeadline(time.Time{})
		}

		switch b[0] {
		case gateway.BytePing:
			select {
			case respCh := <-bc.goCh:
				piped = true // pipe or probeBinderConn takes ownership; don't close on exit
				mod.sendSignal(bc, respCh)
				return
			default:
				if _, err := bc.Conn.Write([]byte{gateway.BytePong}); err != nil {
					return
				}
			}
		default:
			return
		}
	}
}

// sendSignal sends ByteSignalGo to the binder and waits for ByteSignalReady.
// Called from keepalive when a connector signals via goCh.
func (mod *Module) sendSignal(bc *binderConn, respCh chan error) {
	if _, err := bc.Conn.Write([]byte{gateway.ByteSignalGo}); err != nil {
		respCh <- err
		return
	}
	if d, ok := bc.Conn.(deadliner); ok {
		d.SetReadDeadline(time.Now().Add(socketProbeTimeout))
	}
	var b [1]byte
	if _, err := io.ReadFull(bc.Conn, b[:]); err != nil {
		respCh <- err
		return
	}
	if d, ok := bc.Conn.(deadliner); ok {
		d.SetReadDeadline(time.Time{})
	}
	if b[0] != gateway.ByteSignalReady {
		respCh <- fmt.Errorf("expected signalReady, got 0x%02x", b[0])
		return
	}
	respCh <- nil
}

// probeBinderConn signals a binderConn via its ping loop to verify liveness.
// It will try at most socketProbeMaxAttempts connections before giving up.
func (mod *Module) probeBinderConn(b *binder, first *binderConn) *binderConn {
	candidate := first

	for attempts := 0; attempts < socketProbeMaxAttempts; attempts++ {
		if candidate == nil {
			var ok bool
			candidate, ok = b.takeConn()
			if !ok {
				return nil
			}
		}

		respCh := make(chan error, 1)
		select {
		case candidate.goCh <- respCh:
		case <-time.After(socketProbeTimeout):
			// goCh buffer full — another connector is already signaling this conn
			candidate.Close()
			candidate = nil
			continue
		}

		select {
		case err := <-respCh:
			if err != nil {
				// sendSignal failed; keepalive already exited via defer, conn is closed
				candidate.Close()
				candidate = nil
				continue
			}
			return candidate
		case <-candidate.done:
			// keepalive exited without responding — conn was dead
			candidate = nil
			continue
		case <-time.After(socketProbeTimeout + socketPingInterval):
			candidate.Close()
			candidate = nil
		}
	}

	mod.log.Errorv(1, "binder %v probe exhausted", b.Identity)
	return nil
}

type deadliner interface {
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

func pipe(a, b io.ReadWriteCloser) {
	const idle = 30 * time.Second

	done := make(chan struct{}, 2)

	copy := func(dst, src io.ReadWriteCloser) {
		// note: sync.Pool could reduce per-connection allocations under high concurrency (pattern used by nginx, envoy, treafik)
		buf := make([]byte, 32*1024)
		srcD, srcOk := src.(deadliner)
		dstD, dstOk := dst.(deadliner)
		for {
			if srcOk {
				srcD.SetReadDeadline(time.Now().Add(idle))
			}
			n, err := src.Read(buf)
			if n > 0 {
				if dstOk {
					dstD.SetWriteDeadline(time.Now().Add(idle))
				}
				if _, werr := dst.Write(buf[:n]); werr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
		done <- struct{}{}
	}

	go copy(a, b)
	go copy(b, a)

	<-done
	a.Close()
	b.Close()
}
