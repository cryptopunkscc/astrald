package infra

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

func ParseUnsupportedAddr(network string, hexData string) (*UnsupportedAddr, error) {
	data, err := hex.DecodeString(hexData)
	if err != nil {
		return nil, err
	}

	return &UnsupportedAddr{network: network, data: data}, nil
}

func (addr UnsupportedAddr) Network() string {
	return addr.network
}

func (addr UnsupportedAddr) String() string {
	return hex.EncodeToString(addr.data)
}

func (addr UnsupportedAddr) Pack() []byte {
	return addr.data
}
