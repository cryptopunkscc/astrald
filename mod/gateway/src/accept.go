package gateway

import (
	"context"
	"fmt"
	"sync"

	"io"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
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
		mod.log.Infov(1, "added idle conn to binder %v", b.Identity)
		b.addConn(conn)
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

	cc := &connectorConn{
		Conn:    conn,
		network: conn.RemoteEndpoint().Network(),
		pipedTo: reserved,
	}

	targetBinder.markPiped(reserved, cc)

	mod.log.Infov(1, "pipe from %v to %v created", c.Identity, c.Target)
	go pipe(reserved, cc)
	return stopListener, nil
}

func pipe(a, b io.ReadWriteCloser) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(a, b)
		a.Close()
	}()

	go func() {
		defer wg.Done()
		io.Copy(b, a)
		b.Close()
	}()

	wg.Wait()
}
