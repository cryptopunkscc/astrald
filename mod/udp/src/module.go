package udp

import (
	"context"
	"net"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/udp"
	"github.com/cryptopunkscc/astrald/tasks"
)

// Module represents the UDP module and implements the exonet.Dialer interface.
type Module struct {
	Deps
	config          Config // Configuration for the module
	node            astral.Node
	assets          assets.Assets
	log             *log.Logger
	ctx             context.Context
	publicEndpoints []exonet.Endpoint
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	err := tasks.Group(NewServer(mod)).Run(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}

func (mod *Module) ListenPort() int {
	return mod.config.ListenPort
}

func (mod *Module) localIPs() ([]udp.IP, error) {
	list := make([]udp.IP, 0)

	ifaceAddrs, err := InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range ifaceAddrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		ip := udp.IP(ipnet.IP)
		list = append(list, ip)
	}

	return list, nil
}

func (mod *Module) localEndpoints() (list []exonet.Endpoint) {
	ips, err := mod.localIPs()
	if err != nil {
		return
	}

	for _, ip := range ips {
		if ip.IsLoopback() {
			continue
		}
		if ip.IsGlobalUnicast() || ip.IsPrivate() {
			list = append(list, &udp.Endpoint{
				IP:   ip,
				Port: astral.Uint16(mod.ListenPort()),
			})
		}
	}
	return
}
