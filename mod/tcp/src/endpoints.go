package tcp

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	_net "net"
)

var _ node.EndpointLister = &Module{}

func (mod *Module) Endpoints() []net.Endpoint {
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
			list = append(list, Endpoint{ip: ipv4, port: uint16(mod.config.ListenPort)})
		}
	}

	// Add custom addresses
	for _, e := range mod.publicEndpoints {
		list = append(list, e)
	}

	return list
}
