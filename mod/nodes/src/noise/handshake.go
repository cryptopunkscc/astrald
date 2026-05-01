package noise

import (
	"context"
	"errors"
	"io"

	"github.com/cryptopunkscc/astrald/brontide"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// HandshakeInbound performs a handshake as the passive party.
func HandshakeInbound(ctx context.Context, conn io.ReadWriteCloser, localPrivateKey *secp256k1.PrivateKey) (*Conn, error) {
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

	bConn, err := brontide.PassiveHandshake(conn, localPrivateKey)
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	if err != nil {
		return nil, err
	}

	return &Conn{brontide: bConn}, nil
}

// HandshakeOutbound performs a handshake as the active party.
func HandshakeOutbound(
	ctx context.Context,
	conn io.ReadWriteCloser,
	remotePublicKey *secp256k1.PublicKey,
	localPrivateKey *secp256k1.PrivateKey,
) (*Conn, error) {

	if localPrivateKey.PubKey().IsEqual(remotePublicKey) {
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
	c, err := brontide.ActiveHandshake(conn, localPrivateKey, remotePublicKey)
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	if err != nil {
		return nil, err
	}

	return &Conn{brontide: c}, nil
}
