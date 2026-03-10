package gateway

import (
	"context"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	tcpmod "github.com/cryptopunkscc/astrald/mod/tcp"
)

func (mod *Module) startServers(ctx *astral.Context) {
	for network, port := range mod.config.Sockets {
		switch network {
		case "tcp":
			endpoint, err := mod.startTCPServer(ctx, port)
			if err != nil {
				mod.log.Error("start gateway tcp server on port %v: %v", port, err)
				continue
			}
			mod.listenEndpoints.Set(network, endpoint)
		default:
			mod.log.Error("unsupported gateway socket network: %v", network)
		}
	}
}

func (mod *Module) startTCPServer(ctx *astral.Context, port uint16) (*tcpmod.Endpoint, error) {
	server := mod.TCP.NewServer(astral.Uint16(port), mod.acceptConn)

	endpoint := &tcpmod.Endpoint{Port: astral.Uint16(port)}
	ips, _ := mod.IP.LocalIPs()
	for _, ip := range ips {
		if !ip.IsLoopback() {
			endpoint.IP = ip
			break
		}
	}
	if endpoint.IP == nil && len(ips) > 0 {
		endpoint.IP = ips[0]
	}

	go func() {
		if err := server.Run(ctx); err != nil {
			mod.log.Error("gateway tcp server: %v", err)
		}
	}()

	return endpoint, nil
}

func (mod *Module) acceptConn(_ context.Context, conn exonet.Conn) (bool, error) {
	var nonce astral.Nonce
	if _, err := nonce.ReadFrom(conn); err != nil {
		mod.log.Errorv(1, "read nonce from %v: %v", conn.RemoteEndpoint(), err)
		conn.Close()
		return false, nil
	}

	if binder, ok := mod.binderByNonce(nonce); ok {
		mod.log.Infov(1, "slot conn from %v via %v", binder.Identity, conn.RemoteEndpoint())
		binder.ConnPool.add(conn)
		return false, nil
	}

	if connecting, ok := mod.connectingByNonce(nonce); ok {
		mod.connecting.Remove(connecting)

		binder, ok := mod.binderByIdentity(connecting.Target)
		if !ok {
			mod.log.Errorv(1, "no binder for %v", connecting.Target)
			conn.Close()
			return false, nil
		}

		binderConnection, ok := binder.ConnPool.take()
		if !ok {
			mod.log.Errorv(1, "no available binderConnection for %v", connecting.Target)
			conn.Close()
			return false, nil
		}

		mod.log.Infov(1, "connecting %v to %v", connecting.Identity, connecting.Target)

		go func() {
			defer binderConnection.Close()
			defer conn.Close()
			pipe(binderConnection, conn)
		}()
		return false, nil
	}

	mod.log.Errorv(1, "unknown nonce %v from %v", nonce, conn.RemoteEndpoint())
	conn.Close()
	return false, nil
}

func pipe(a, b io.ReadWriteCloser) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		io.Copy(a, b)
	}()
	io.Copy(b, a)
	<-done
}
