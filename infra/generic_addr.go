package infra

import (
	"encoding/hex"
)

type GenericAddr struct {
	network string
	bytes   []byte
}

func NewGenericAddr(network string, bytes []byte) *GenericAddr {
	return &GenericAddr{network: network, bytes: bytes}
}

func (g *GenericAddr) Network() string {
	return g.network
}

func (g *GenericAddr) String() string {
	return hex.EncodeToString(g.bytes)
}

func (g *GenericAddr) Pack() []byte {
	return g.bytes
}
