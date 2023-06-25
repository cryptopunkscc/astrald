package net

import (
	"bytes"
	"encoding/hex"
)

// Endpoint represents a dialable address on a network (such as IP address and port)
type Endpoint interface {
	Network() string // network name
	String() string  // text representation of the address
	Pack() []byte    // binary represenation of the address
}

// EndpointEqual compares two addresses
func EndpointEqual(a, b Endpoint) bool {
	if a.Network() != b.Network() {
		return false
	}

	return bytes.Equal(a.Pack(), b.Pack())
}

type GenericEndpoint struct {
	network string
	bytes   []byte
}

func NewGenericEndpoint(network string, bytes []byte) *GenericEndpoint {
	return &GenericEndpoint{network: network, bytes: bytes}
}

func (g *GenericEndpoint) Network() string {
	return g.network
}

func (g *GenericEndpoint) String() string {
	return hex.EncodeToString(g.bytes)
}

func (g *GenericEndpoint) Pack() []byte {
	return g.bytes
}
