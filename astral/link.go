package astral

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
)

// Link attempts to establish a link with a remote identity over the provided transport.
func Link(ctx context.Context, localID id.Identity, remoteID id.Identity, conn infra.Conn) (*link.Link, error) {
	// make sure we're not linking with ourselves
	if localID.IsEqual(remoteID) {
		return nil, errors.New("localID and remoteID cannot be equal")
	}

	authConn, err := auth.HandshakeOutbound(ctx, conn, remoteID, localID)
	if err != nil {
		return nil, err
	}

	return link.New(authConn), nil
}

// LinkFirst iterates over connCh and attempts to establish a link over each connection. It returns the first
// successfully establlished link and closes all subsequest connections from connCh.
func LinkFirst(ctx context.Context, localID id.Identity, remoteID id.Identity, connCh <-chan infra.Conn) *link.Link {
	// at the end, close all remaining connections
	defer func() {
		for conn := range connCh {
			conn.Close()
		}
	}()

	// go through conns until link is established
	for conn := range connCh {
		if link, err := Link(ctx, localID, remoteID, conn); err == nil {
			return link
		}

		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}

	return nil
}
