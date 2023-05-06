package auth

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/brontide"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

// HandshakeInbound performs a handshake as the passive party.
func HandshakeInbound(ctx context.Context, conn net.Conn, localID id.Identity) (*NoiseConn, error) {
	//TODO: is there a better way to handle ctx here?
	var done = make(chan struct{})
	var errCh = make(chan error, 1)
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			conn.Close()
		case <-done:
		}
	}()

	bConn, err := brontide.PassiveHandshake(conn, localID.PrivateKey())
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	if err != nil {
		return nil, err
	}

	return &NoiseConn{
		conn:     conn,
		brontide: bConn,
	}, nil
}

// HandshakeOutbound performs a handshake as the active party.
func HandshakeOutbound(ctx context.Context, conn net.Conn, expectedRemoteID id.Identity, localID id.Identity) (*NoiseConn, error) {
	//TODO: is there a better way to handle ctx here?
	var done = make(chan struct{})
	var errCh = make(chan error, 1)
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			conn.Close()
		case <-done:
		}
	}()
	c, err := brontide.ActiveHandshake(conn, localID.PrivateKey(), expectedRemoteID.PublicKey())
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	if err != nil {
		return nil, err
	}

	return &NoiseConn{
		conn:     conn,
		brontide: c,
	}, nil
}
