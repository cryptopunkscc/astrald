package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
)

// Infra is an interface for infrastructural networks
type Infra interface {
	Dial(ctx context.Context, addr net.Endpoint) (conn net.Conn, err error)
	Unpack(network string, data []byte) (net.Endpoint, error)
	Parse(network string, address string) (net.Endpoint, error)
	Endpoints() []net.Endpoint
	SetDialer(network string, dialer Dialer)
	SetParser(network string, parser Parser)
	SetUnpacker(network string, unpacker Unpacker)
	AddEndpoints(e EndpointLister) error
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
