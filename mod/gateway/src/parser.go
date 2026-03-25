package gateway

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

var _ exonet.Parser = &Module{}

func (mod *Module) Parse(network string, address string) (exonet.Endpoint, error) {
	if network != NetworkName {
		return nil, exonet.ErrUnsupportedNetwork
	}

	var ids = strings.SplitN(address, ":", 2)
	if len(ids) != 2 {
		return nil, fmt.Errorf("invalid endpoint: %s", address)
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
