package linker

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/link"
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
