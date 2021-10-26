package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
)

type Network interface {
	Name() string
	Unpack([]byte) (Addr, error)

	Dial(ctx context.Context, addr Addr) (Conn, error)
	Listen(ctx context.Context) (<-chan Conn, <-chan error)

	Broadcast(ctx context.Context, payload []byte) <-chan error
	Scan(ctx context.Context) (<-chan Broadcast, <-chan error)

	//	Announce(ctx context.Context, id id.Identity) error
	//	Discover(ctx context.Context) <-chan Presence

	Addresses() []AddrDesc
}

type Presence struct {
	Identity id.Identity
	Addr     Addr
}

type Broadcast struct {
	SourceAddr Addr
	Payload    []Addr
}
