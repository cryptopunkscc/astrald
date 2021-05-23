package net

import (
	"net"
)

type Addr net.Addr

// address represents an abstract addressable point in a network.
type address struct {
	network string
	address string
}

func MakeAddr(net, addr string) Addr {
	return &address{
		network: net,
		address: addr,
	}
}

var _ Addr = address{}

func (a address) Network() string {
	return a.network
}

func (a address) String() string {
	return a.address
}
