package infra

import "context"

// Broadcast holds information about an incoming broadcast
type Broadcast struct {
	SourceAddr Addr
	Payload    []Addr
}

// Broadcaster wraps the Broadcast method. Broadcast sends a payload to everyone on the network.
type Broadcaster interface {
	Broadcast(payload []byte) error
}

// Scanner wraps the Scan method. Scan listens for any broadcasts on the network.
type Scanner interface {
	Scan(ctx context.Context) (<-chan Broadcast, error)
}

// BroadcastNet combines interfaces specific to broadcast networks
type BroadcastNet interface {
	Network
	Broadcaster
	Scanner
}
