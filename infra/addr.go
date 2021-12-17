package infra

import (
	"bytes"
)

type Addr interface {
	Network() string // network name
	String() string  // string representation of the address for display
	Pack() []byte    // network-specific binary represenation of the address
}

// Unpacker wraps the Unpack method. Unpack deserializes network-specific binary data into an Addr struct.
type Unpacker interface {
	Unpack(network string, data []byte) (Addr, error)
}

// AddrSpec holds additional information about an Addr.
type AddrSpec struct {
	Addr
	Global bool
}

// AddrEqual compares two addresses
func AddrEqual(a, b Addr) bool {
	if a.Network() != b.Network() {
		return false
	}

	return bytes.Equal(a.Pack(), b.Pack())
}
