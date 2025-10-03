package utp

import (
	"bytes"

	"github.com/cryptopunkscc/astrald/mod/exonet"

	"github.com/cryptopunkscc/astrald/mod/utp"
)

var _ exonet.Unpacker = &Module{}

func (mod *Module) Unpack(network string, data []byte) (exonet.Endpoint, error) {
	switch network {
	case "utp":
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}
	return Unpack(data)
}

func Unpack(buf []byte) (e *utp.Endpoint, err error) {
	e = &utp.Endpoint{}
	_, err = e.ReadFrom(bytes.NewReader(buf))
	return
}
