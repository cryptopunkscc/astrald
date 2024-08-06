package noise

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/brontide"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

// HandshakeInbound performs a handshake as the passive party.
func HandshakeInbound(ctx context.Context, conn exonet.Conn, localID *astral.Identity) (*Conn, error) {
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

	return &Conn{
		conn:     conn,
		brontide: bConn,
	}, nil
}

// HandshakeOutbound performs a handshake as the active party.
func HandshakeOutbound(ctx context.Context, conn exonet.Conn, expectedRemoteID *astral.Identity, localID *astral.Identity) (*Conn, error) {
	if localID.IsEqual(expectedRemoteID) {
		return nil, errors.New("local and remote identities cannot be equal")
	}

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

	return &Conn{
		conn:     conn,
		brontide: c,
	}, nil
}
