package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
)

type Network interface {
	// Name returns the name of the network
	Name() string

	// Unpack parses a binary representation of an address
	Unpack([]byte) (Addr, error)

	// Dial opens an unicast connection with the addresss
	Dial(ctx context.Context, addr Addr) (Conn, error)

	// Listen starts accepting incoming unicast connections
	Listen(ctx context.Context) (<-chan Conn, <-chan error)

	// Broadcast sends a payload to everyone on the network
	Broadcast(payload []byte) error

	// Scan scans the network for broadcasts
	Scan(ctx context.Context) (<-chan Broadcast, <-chan error)

	// Announce announces identity's presence on the network
	Announce(ctx context.Context, id id.Identity) error

	// Discover discovers other peers on the network
	Discover(ctx context.Context) (<-chan Presence, error)

	// Addresses returns a list of our addresses on this network
	Addresses() []AddrDesc
}

// Presence holds information about an identity present on the network
type Presence struct {
	Identity id.Identity
	Addr     Addr
}

// Broadcast holds information about an incoming broadcast
type Broadcast struct {
	SourceAddr Addr
	Payload    []Addr
}
