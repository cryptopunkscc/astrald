package inet

import (
	"github.com/cryptopunkscc/astrald/infra"
	"log"
	"net"
)

var _ infra.Network = &Inet{}

type Inet struct {
	config         Config
	listenPort     uint16
	publicAddrs    []infra.AddrSpec
	separateListen bool
}

func New(config Config) *Inet {
	inet := &Inet{
		config:      config,
		listenPort:  defaultListenPort,
		publicAddrs: make([]infra.AddrSpec, 0),
	}

	// Add public addresses
	for _, addrStr := range config.PublicAddr {
		addr, err := Parse(addrStr)
		if err != nil {
			log.Println("inet: parse error:", err)
			continue
		}

		inet.publicAddrs = append(inet.publicAddrs, infra.AddrSpec{
			Addr:   addr,
			Global: true,
		})
		log.Println("inet: added", addr)
	}

	return inet
}

func (inet Inet) Name() string {
	return NetworkName
}

func (inet Inet) Addresses() []infra.AddrSpec {
	list := make([]infra.AddrSpec, 0)

	ifaceAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	for _, a := range ifaceAddrs {
		ipnet, ok := a.(*net.IPNet)
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
			list = append(list, infra.AddrSpec{
				Addr:   Addr{ip: ipv4, port: inet.listenPort},
				Global: ipv4.IsGlobalUnicast() && (!ipv4.IsPrivate()),
			})
		}
	}

	// Add custom addresses
	list = append(list, inet.publicAddrs...)

	return list
}
