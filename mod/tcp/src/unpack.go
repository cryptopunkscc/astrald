package tcp

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
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

func Unpack(buf []byte) (addr *Endpoint, err error) {
	var r = bytes.NewReader(buf)

	addr = &Endpoint{}

	if err = cslq.Decode(r, "c", &addr.ver); err != nil {
		return
	}

	switch addr.ver {
	case ipv4:
		return addr, cslq.Decode(r, "[4]c s", &addr.ip, &addr.port)
	case ipv6:
		return addr, cslq.Decode(r, "[16]c s", &addr.ip, &addr.port)
	}

	return addr, errors.New("invalid version")
}
