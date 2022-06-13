package astral

import (
	"net"
)

var _ net.Addr = &Addr{}

type Addr struct {
	address string
}

func (a Addr) Network() string {
	return "astral"
}

func (a Addr) String() string {
	return a.address
}
