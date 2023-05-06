package net

import (
	"bytes"
	"encoding/hex"
)

type Endpoint interface {
	Network() string // network name
	String() string  // string representation of the address for display
	Pack() []byte    // network-specific binary represenation of the address
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
