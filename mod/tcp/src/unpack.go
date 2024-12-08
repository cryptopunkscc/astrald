package tcp

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

var _ exonet.Unpacker = &Module{}

func (mod *Module) Unpack(network string, data []byte) (exonet.Endpoint, error) {
	switch network {
	case "tcp", "inet":
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}
	return Unpack(data)
}

func Unpack(buf []byte) (e *Endpoint, err error) {
	e = &Endpoint{}
	_, err = e.ReadFrom(bytes.NewReader(buf))
	return
}
