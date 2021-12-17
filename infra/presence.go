package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
)

// Presence holds information about an identity present on the network
type Presence struct {
	Identity id.Identity
	Addr     Addr
}

// Announcer wraps the Announce method. Announce announces node's presence on the network using the provided identity.
type Announcer interface {
	Announce(ctx context.Context, id id.Identity) error
}

// Discoverer wraps the Discover method. Discover discovers presence of other peers on the network.
type Discoverer interface {
	Discover(ctx context.Context) (<-chan Presence, error)
}

// PresenceNet combines interfaces specific to presence networks
type PresenceNet interface {
	Announcer
	Discoverer
}
