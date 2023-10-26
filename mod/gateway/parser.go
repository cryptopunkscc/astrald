package gateway

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	"strings"
)

var _ infra.Parser = &Module{}

func (mod *Module) Parse(network string, address string) (net.Endpoint, error) {
	if network != NetworkName {
		return nil, infra.ErrUnsupportedNetwork
	}

	var ids = strings.SplitN(address, ":", 2)
	if len(ids) != 2 {
		return nil, ErrParseError{msg: "invalid address string"}
	}

	var err error
	var endpoint Endpoint

	endpoint.gate, err = mod.node.Resolver().Resolve(ids[0])
	if err != nil {
		return nil, err
	}
	endpoint.target, err = mod.node.Resolver().Resolve(ids[1])
	if err != nil {
		return nil, err
	}

	if endpoint.gate.IsEqual(endpoint.target) {
		return nil, errors.New("invalid endpoint")
	}

	return endpoint, nil
}

// Parse converts a text representation of a gateway address to an Endpoint struct
func Parse(str string) (endpoint Endpoint, err error) {
	if len(str) != (2*66)+1 { // two public key hex strings and a separator ":"
		return endpoint, ErrParseError{msg: "invalid address length"}
	}
	var ids = strings.SplitN(str, ":", 2)
	if len(ids) != 2 {
		return Endpoint{}, ErrParseError{msg: "invalid address string"}
	}
	endpoint.gate, err = id.ParsePublicKeyHex(ids[0])
	if err != nil {
		return Endpoint{}, err
	}
	endpoint.target, err = id.ParsePublicKeyHex(ids[1])
	if err != nil {
		return Endpoint{}, err
	}
	if endpoint.gate.IsEqual(endpoint.target) {
		return Endpoint{}, errors.New("invalid endpoint")
	}

	return
}
