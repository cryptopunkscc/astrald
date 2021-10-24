package infra

import "bytes"

type Addr interface {
	Network() string // network name
	String() string  // string representation of the address for display
	Pack() []byte    // serialized binary address data
}

type AddrDesc struct {
	Addr
	Public bool
}

func AddrEqual(a, b Addr) bool {
	if a.Network() != b.Network() {
		return false
	}

	return bytes.Equal(a.Pack(), b.Pack())
}
