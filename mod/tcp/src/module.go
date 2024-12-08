package tcp

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/tasks"
	"net"
)

var _ tcp.Module = &Module{}

type Module struct {
	Deps
	config          Config
	node            astral.Node
	log             *log.Logger
	ctx             context.Context
	publicEndpoints []exonet.Endpoint
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	tasks.Group(NewServer(mod)).Run(ctx)

	<-ctx.Done()

	return nil
}

func (mod *Module) Resolve(ctx context.Context, identity *astral.Identity) (endpoints []exonet.Endpoint, err error) {
	if !identity.IsEqual(mod.node.Identity()) {
		return
	}

	endpoints = append(endpoints, mod.publicEndpoints...)
	endpoints = append(endpoints, mod.localEndpoints()...)

	return
}

func (mod *Module) ListenPort() int {
	return mod.config.ListenPort
}

func (mod *Module) localIPs() ([]tcp.IP, error) {
	list := make([]tcp.IP, 0)

	ifaceAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range ifaceAddrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		ip := tcp.IP(ipnet.IP)
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
			list = append(list, &Endpoint{
				ip:   ip,
				port: astral.Uint16(mod.ListenPort()),
			})
		}
	}
	return
}
