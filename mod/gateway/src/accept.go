package gateway

import (
	"context"

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
func (mod *Module) acceptSocketConn(_ context.Context, conn exonet.Conn) (bool, error) {
	mod.log.Infov(1, "accepting connection from %v", conn.RemoteEndpoint())

	var nonce astral.Nonce
	if _, err := nonce.ReadFrom(conn); err != nil {
		mod.log.Errorv(1, "read nonce from %v: %v", conn.RemoteEndpoint(), err)
		conn.Close()
		return true, nil
	}

	client, ok := mod.clientByNonce(nonce)
	if !ok {
		mod.log.Errorv(1, "unknown nonce %v from %v", nonce, conn.RemoteEndpoint())
		conn.Close()
		return true, nil
	}

	if client.isBinder() {
		mod.log.Infov(1, "added idle conn to %v", client.Identity)
		client.add(conn)
		return false, nil
	}

	// connecting
	mod.connecting.Remove(client)
	binderConn := client.takePipeTo()
	if binderConn == nil {
		mod.log.Errorv(1, "no reserved conn for %v", client.Target)
		conn.Close()
		return true, nil
	}

	connectorConn := &clientConn{
		Conn:    conn,
		network: conn.RemoteEndpoint().Network(),
	}

	client.conns.Add(connectorConn)

	targetClient, ok := mod.binderByIdentity(client.Target)
	if !ok {
		return true, nil
	}

	targetClient.markPiped(binderConn, connectorConn)
	client.markPiped(connectorConn, binderConn)

	mod.log.Infov(1, "connecting %v to %v", client.Identity, client.Target)
	go pipe(binderConn, connectorConn)
	return false, nil
}

func pipe(a, b io.ReadWriteCloser) {
	done := make(chan struct{}, 2)
	defer close(done)

	go func() {
		io.Copy(a, b)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(b, a)
		done <- struct{}{}
	}()

	<-done
	a.Close()
	b.Close()
}
