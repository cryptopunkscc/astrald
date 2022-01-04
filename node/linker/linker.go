package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
)

// Linker is an interface that wraps Link method. Link tries to produce and return a new link. Returns nil on error.
type Linker interface {
	Link(ctx context.Context, remoteID id.Identity) *link.Link
}
