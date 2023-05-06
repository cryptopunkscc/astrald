package gw

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	"strings"
)

var _ infra.Parser = &Driver{}

func (drv *Driver) Parse(network string, address string) (net.Endpoint, error) {
	return Parse(address)
}

// Parse converts a text representation of a gateway address to an Endpoint struct
func Parse(str string) (addr Endpoint, err error) {
	if len(str) != (2*66)+1 { // two public key hex strings and a separator ":"
		return addr, errors.New("invalid address length")
	}
	var ids = strings.SplitN(str, ":", 2)
	if len(ids) != 2 {
		return addr, errors.New("invalid address string")
	}
	addr.gate, err = id.ParsePublicKeyHex(ids[0])
	if err != nil {
		return
	}
	addr.target, err = id.ParsePublicKeyHex(ids[1])
	return
}
