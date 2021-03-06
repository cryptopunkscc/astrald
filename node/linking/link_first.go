package linking

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/link"
)

// LinkFirst iterates over connCh and attempts to establish a link over each connection. It returns the first
// successfully establlished link and closes all subsequest connections from connCh.
func LinkFirst(ctx context.Context, localID id.Identity, remoteID id.Identity, connCh <-chan infra.Conn) *link.Link {
	if localID.IsEqual(remoteID) {
		return nil
	}

	// at the end, close all remaining connections
	defer func() {
		// do this in a new routine, otherwise return will hang until connCh is closed
		go func() {
			for conn := range connCh {
				conn.Close()
			}
		}()
	}()

	// go through conns until link is established
	for conn := range connCh {
		if link, err := Link(ctx, localID, remoteID, conn); err == nil {
			return link
		} else {
			if errors.Is(err, context.Canceled) {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			break
		default:
		}
	}

	return nil
}
