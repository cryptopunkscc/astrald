package gateway

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"strings"
)

var _ exonet.Parser = &Module{}

func (mod *Module) Parse(network string, address string) (exonet.Endpoint, error) {
	if network != NetworkName {
		return nil, exonet.ErrUnsupportedNetwork
	}

	var ids = strings.SplitN(address, ":", 2)
	if len(ids) != 2 {
		return nil, ErrParseError{msg: "invalid address string"}
	}

	var err error
	var endpoint gateway.Endpoint

	endpoint.GatewayID, err = mod.Dir.ResolveIdentity(ids[0])
	if err != nil {
		return nil, err
	}
	endpoint.TargetID, err = mod.Dir.ResolveIdentity(ids[1])
	if err != nil {
		return nil, err
	}

	if endpoint.GatewayID.IsEqual(endpoint.TargetID) {
		return nil, errors.New("invalid endpoint")
	}

	return &endpoint, nil
}

// Parse converts a text representation of a gateway address to an Endpoint struct
func Parse(str string) (endpoint *gateway.Endpoint, err error) {
	if len(str) != (2*66)+1 { // two public key hex strings and a separator ":"
		return endpoint, ErrParseError{msg: "invalid address length"}
	}
	var ids = strings.SplitN(str, ":", 2)
	if len(ids) != 2 {
		return nil, ErrParseError{msg: "invalid address string"}
	}
	endpoint.GatewayID, err = astral.ParseIdentity(ids[0])
	if err != nil {
		return nil, err
	}
	endpoint.TargetID, err = astral.ParseIdentity(ids[1])
	if err != nil {
		return nil, err
	}
	if endpoint.GatewayID.IsEqual(endpoint.TargetID) {
		return nil, errors.New("invalid endpoint")
	}

	return
}
