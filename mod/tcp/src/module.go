package tcp

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/tasks"
	"net"
	"strings"
	"time"
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

	go mod.watchAddresses(ctx)

	tasks.Group(NewServer(mod)).Run(ctx)

	<-ctx.Done()

	return nil
}

func (mod *Module) ResolveEndpoints(ctx context.Context, identity *astral.Identity) (endpoints []exonet.Endpoint, err error) {
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

	ifaceAddrs, err := InterfaceAddrs()
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
			list = append(list, &tcp.Endpoint{
				IP:   ip,
				Port: astral.Uint16(mod.ListenPort()),
			})
		}
	}
	return
}

func (mod *Module) watchAddresses(ctx context.Context) {
	_, err := InterfaceAddrs()
	if err != nil {
		mod.log.Errorv(0, "network interface monitoring disabled: %v", err)
		return
	}
	addrs := mod.getAddresses()
	for {
		select {
		case <-time.After(3 * time.Second):
			newAddrs := mod.getAddresses()
			removed, added := diff(addrs, newAddrs)
			addrs = newAddrs

			if len(removed) > 0 || len(added) > 0 {
				mod.log.Logv(1, "network addresses changed. added: %v; removed: %v",
					strings.Join(added, ", "),
					strings.Join(removed, ", "),
				)
				mod.Objects.Receive(&tcp.EventNetworkAddressChanged{
					Removed: removed,
					Added:   added,
					All:     addrs,
				}, nil)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (mod *Module) getAddresses() (s []string) {
	addrs, _ := InterfaceAddrs()
	for _, addr := range addrs {
		s = append(s, addr.String())
	}
	return
}

func diff(a, b []string) ([]string, []string) {
	// Since the arrays are ordered, we can use two pointers to compare elements
	var onlyInA, onlyInB []string
	i, j := 0, 0

	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			i++
			j++
		case a[i] < b[j]:
			// 'a[i]' is not in 'b'
			onlyInA = append(onlyInA, a[i])
			i++
		case a[i] > b[j]:
			// 'b[j]' is not in 'a'
			onlyInB = append(onlyInB, b[j])
			j++
		}
	}

	// Append any remaining elements from 'a'
	for i < len(a) {
		onlyInA = append(onlyInA, a[i])
		i++
	}

	// Append any remaining elements from 'b'
	for j < len(b) {
		onlyInB = append(onlyInB, b[j])
		j++
	}

	return onlyInA, onlyInB
}
