package exonet

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "exonet"

type Module interface {
	Dial(context.Context, Endpoint) (conn Conn, err error)
	Unpack(network string, data []byte) (Endpoint, error)
	Parse(network string, address string) (Endpoint, error)
	ResolveEndpoints(context.Context, *astral.Identity) ([]Endpoint, error)

	SetDialer(network string, dialer Dialer)
	SetUnpacker(network string, unpacker Unpacker)
	SetParser(network string, parser Parser)
	AddResolver(EndpointResolver)
}

// Endpoint represents a dialable address on a network (such as an IP address with port number)
type Endpoint interface {
	Network() string // network name
	Address() string // text representation of the address
	Pack() []byte    // binary represenation of the address
}

type Dialer interface {
	Dial(context.Context, Endpoint) (Conn, error)
}

type Unpacker interface {
	Unpack(network string, data []byte) (Endpoint, error)
}

type Parser interface {
	Parse(network string, address string) (Endpoint, error)
}

type EndpointResolver interface {
	ResolveEndpoints(context.Context, *astral.Identity) ([]Endpoint, error)
}
