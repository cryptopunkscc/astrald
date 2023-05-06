package bt

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	"strconv"
	"strings"
)

var _ infra.Parser = &Driver{}

func (drv *Driver) Parse(network string, address string) (net.Endpoint, error) {
	return Parse(address)
}

func Parse(addr string) (Endpoint, error) {
	a := strings.Split(addr, ":")
	if len(a) != 6 {
		return Endpoint{}, infra.ErrInvalidAddress
	}

	var mac [6]byte

	for i, b := range a {
		u, err := strconv.ParseUint(b, 16, 8)
		if err != nil {
			return Endpoint{}, infra.ErrInvalidAddress
		}
		mac[len(mac)-1-i] = byte(u)
	}

	return Endpoint{mac: mac}, nil
}
