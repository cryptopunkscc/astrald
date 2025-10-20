package kcp

import (
	"bytes"

	"github.com/cryptopunkscc/astrald/mod/exonet"
	kcpmod "github.com/cryptopunkscc/astrald/mod/kcp"
)

var _ exonet.Unpacker = &Module{}

func (mod *Module) Unpack(network string, data []byte) (exonet.Endpoint, error) {
	switch network {
	case "kcp":
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}
	return Unpack(data)
}

func Unpack(buf []byte) (e *kcpmod.Endpoint, err error) {
	e = &kcpmod.Endpoint{}
	_, err = e.ReadFrom(bytes.NewReader(buf))
	return
}
