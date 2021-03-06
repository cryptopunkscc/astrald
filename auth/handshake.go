package auth

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/brontide"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
)

// HandshakeInbound performs a handshake as the passive party.
func HandshakeInbound(ctx context.Context, conn infra.Conn, localID id.Identity) (Conn, error) {
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
func HandshakeOutbound(ctx context.Context, conn infra.Conn, expectedRemoteID id.Identity, localID id.Identity) (Conn, error) {
	c, err := brontide.ActiveHandshake(conn, localID.PrivateKey(), expectedRemoteID.PublicKey())
	if err != nil {
		return nil, err
	}

	return &brontideConn{
		netConn: conn,
		bConn:   c,
	}, nil
}
