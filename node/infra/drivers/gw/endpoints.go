package gw

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
)

var _ infra.EndpointLister = &Driver{}

func (drv *Driver) Endpoints() []net.Endpoint {
	endpoints := make([]net.Endpoint, 0)

	for _, gate := range drv.config.Gateways {
		gateID, err := id.ParsePublicKeyHex(gate)
		if err != nil {
			drv.log.Error("error parsing gateway %s: %s", gate, err.Error())
			continue
		}

		endpoints = append(
			endpoints,
			NewEndpoint(gateID, drv.infra.Node().Identity()),
		)
	}

	return endpoints
}
