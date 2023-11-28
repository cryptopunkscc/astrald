package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/resolver"
	"github.com/cryptopunkscc/astrald/node/router"
)

// Infra is an interface for infrastructural networks
type Infra interface {
	Node() Node
	Dial(ctx context.Context, addr net.Endpoint) (conn net.Conn, err error)
	Endpoints() []net.Endpoint
	Unpack(network string, data []byte) (net.Endpoint, error)
	Parse(network string, address string) (net.Endpoint, error)
	AddDriver(name string, driver Driver) error
	Drivers() map[string]Driver
}

// Node is a subset of node.Node interface with methods that should be exposed to the network drivers
type Node interface {
	Resolver() resolver.Resolver
	Identity() id.Identity
	Router() router.Router
}

// Dialer wraps the Dial method. Dial opens an unicast connection with the provided address.
type Dialer interface {
	Dial(ctx context.Context, addr net.Endpoint) (net.Conn, error)
}

// Listener wraps the Listen method. Listen starts accepting incoming unicast connections.
type Listener interface {
	Listen(ctx context.Context) (<-chan net.Conn, error)
}

type Unpacker interface {
	Unpack(network string, data []byte) (net.Endpoint, error)
}

type Parser interface {
	Parse(network string, address string) (net.Endpoint, error)
}

type EndpointLister interface {
	Endpoints() []net.Endpoint
}
