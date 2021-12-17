package infra

import "context"

// Dialer wraps the Dial method. Dial opens an unicast connection with the provided address.
type Dialer interface {
	Dial(ctx context.Context, addr Addr) (Conn, error)
}

// Listener wraps the Listen method. Listen starts accepting incoming unicast connections.
type Listener interface {
	Listen(ctx context.Context) (<-chan Conn, error)
}

// UnicastNet combines interfaces specific to unicast networks
type UnicastNet interface {
	Network
	Dialer
	Listener
}
