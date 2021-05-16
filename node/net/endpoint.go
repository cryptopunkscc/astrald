package net

import "net"

// Endpoint represents an abstract addressable point in a network.
type Endpoint struct {
	Net     string
	Address string
}

var _ net.Addr = Endpoint{}

func (e Endpoint) Network() string {
	return e.Net
}

func (e Endpoint) String() string {
	return e.Address
}
