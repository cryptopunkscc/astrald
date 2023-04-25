package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"net"
	"sync"
)

var _ infra.Network = &Inet{}

type Inet struct {
	config      Config
	listenPort  int
	publicAddrs []infra.AddrSpec
	mu          sync.Mutex
	localID     id.Identity
}

func New(config Config, localID id.Identity) (*Inet, error) {
	inet := &Inet{
		config:      config,
		publicAddrs: make([]infra.AddrSpec, 0),
		localID:     localID,
	}

	// Add public addresses
	for _, addrStr := range config.PublicAddr {
		addr, err := Parse(addrStr)
		if err != nil {
			log.Error("parse error: %s", err)
			continue
		}

		inet.publicAddrs = append(inet.publicAddrs, infra.AddrSpec{
			Addr:   addr,
			Global: true,
		})
		log.Log("public addr: %s", addr)
	}

	return inet, nil
}

func (inet *Inet) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (inet *Inet) Name() string {
	return NetworkName
}

func (inet *Inet) Addresses() []infra.AddrSpec {
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
				Addr:   Addr{ip: ipv4, port: uint16(inet.ListenPort())},
				Global: ipv4.IsGlobalUnicast() && (!ipv4.IsPrivate()),
			})
		}
	}

	// Add custom addresses
	list = append(list, inet.publicAddrs...)

	return list
}

func (inet *Inet) ListenPort() int {
	if inet.config.ListenPort != 0 {
		return inet.config.ListenPort
	}

	return defaultListenPort
}
