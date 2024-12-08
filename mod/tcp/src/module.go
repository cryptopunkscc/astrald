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
	endpoints = append(endpoints, mod.scanLocalEndpoints()...)

	return
}

func (mod *Module) ListenPort() int {
	return mod.config.ListenPort
}

func (mod *Module) scanLocalEndpoints() []exonet.Endpoint {
	list := make([]exonet.Endpoint, 0)

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
			list = append(list, &Endpoint{ip: tcp.IP(ipv4), port: astral.Uint16(mod.config.ListenPort)})
		}
	}

	return list
}
