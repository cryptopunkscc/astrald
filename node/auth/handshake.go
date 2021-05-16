package auth

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/auth/brontide"
	"github.com/cryptopunkscc/astrald/node/net"
)

// HandshakeInbound performs a handshake as the passive party.
func HandshakeInbound(ctx context.Context, conn net.Conn, localID *ECIdentity) (Conn, error) {
	bConn, err := brontide.PassiveHandshake(conn, localID.PrivateKey())
	if err != nil {
		return nil, err
	}

	return &brontideConn{
		netConn: conn,
		bConn:   bConn,
	}, nil
}

// HandshakeOutbound performs a handshake as the active party.
func HandshakeOutbound(ctx context.Context, conn net.Conn, expectedRemoteID *ECIdentity, localID *ECIdentity) (Conn, error) {
	c, err := brontide.ActiveHandshake(conn, localID.PrivateKey(), expectedRemoteID.PublicKey())
	if err != nil {
		return nil, err
	}

	return &brontideConn{
		netConn: conn,
		bConn:   c,
	}, nil
}
