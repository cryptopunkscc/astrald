package inet

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	_net "net"
)

var _ infra.EndpointLister = &Driver{}

func (drv *Driver) Endpoints() []net.Endpoint {
	list := make([]net.Endpoint, 0)

	ifaceAddrs, err := _net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	for _, a := range ifaceAddrs {
		ipnet, ok := a.(*_net.IPNet)
		if !ok {
			continue
		}

		ipv4 := ipnet.IP.To4()
		if ipv4 == nil {
			continue
		}

		if ipv4.IsLoopback() {
			continue
		}

		if ipv4.IsGlobalUnicast() || ipv4.IsPrivate() {
			list = append(list, Endpoint{ip: ipv4, port: uint16(drv.ListenPort())})
		}
	}

	// Add custom addresses
	list = append(list, drv.publicAddrs...)

	return list
}
