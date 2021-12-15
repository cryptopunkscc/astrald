package node

import (
	"encoding/hex"
	"github.com/cryptopunkscc/astrald/infra"
)

var _ infra.Addr = &UnsupportedAddr{}

type UnsupportedAddr struct {
	network string
	data    []byte
}

func NewUnsupportedAddr(network string, data []byte) *UnsupportedAddr {
	return &UnsupportedAddr{network: network, data: data}
}

func (addr UnsupportedAddr) Network() string {
	return addr.network
}

func (addr UnsupportedAddr) String() string {
	return addr.Network() + ":" + hex.EncodeToString(addr.data)
}

func (addr UnsupportedAddr) Pack() []byte {
	return addr.data
}
