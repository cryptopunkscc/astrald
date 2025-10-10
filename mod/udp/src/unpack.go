package udp

import (
	"bytes"

	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/udp"
)

var _ exonet.Unpacker = &Module{}

func (mod *Module) Unpack(network string, data []byte) (exonet.Endpoint, error) {
	switch network {
	case "udp":
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}
	return Unpack(data)
}

func Unpack(buf []byte) (e *udp.Endpoint, err error) {
	e = &udp.Endpoint{}
	_, err = e.ReadFrom(bytes.NewReader(buf))
	return
}
