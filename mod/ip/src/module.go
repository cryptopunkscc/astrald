package ip

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/resources"
)

var _ ip.Module = &Module{}

type Deps struct {
	Objects objects.Module
}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	assets resources.Resources

	publicIPs []ip.IP
}

func (mod *Module) Run(ctx *astral.Context) error {
	<-ctx.Done()

	return nil
}

func (mod *Module) LocalIPs() ([]ip.IP, error) {
	list := make([]ip.IP, 0)

	ifaceAddrs, err := InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range ifaceAddrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ip.IP(ipnet.IP)
		list = append(list, ip)
	}

	return list, nil
}

func (mod *Module) PublicIPs() []ip.IP {
	return mod.publicIPs
}

func (mod *Module) watchAddresses(ctx context.Context) {
	addrs, err := mod.getAddresses()
	if err != nil {
		mod.log.Errorv(0,
			`ip module/watchAddresses network interface monitoring disabled
%v`, err)
		return
	}

	for {
		select {
		case <-time.After(3 * time.Second):
			newAddrs, err := mod.getAddresses()
			if err != nil {
				mod.log.Errorv(0,
					"ip module/watchAddresses failed to get network addresses"+
						": %v", err)
			}

			removed, added := diff(addrs, newAddrs)
			addrs = newAddrs

			if len(removed) > 0 || len(added) > 0 {
				mod.log.Logv(1, "network addresses changed. added: %v; removed: %v",
					joinIPs(added),
					joinIPs(removed),
				)
				_ = mod.Objects.Receive(&ip.EventNetworkAddressChanged{
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

func (mod *Module) getAddresses() (out []ip.IP, err error) {
	ifaceAddrs, err := InterfaceAddrs()
	if err != nil {
		return out, fmt.Errorf(`ip module/getAddresses: %w`, err)
	}

	for _, a := range ifaceAddrs {
		switch v := a.(type) {
		case *net.IPNet:
			if v.IP == nil {
				continue
			}

			addr := ip.IP(v.IP)

			out = append(out, addr)
		}
	}

	// Ensure deterministic order for diff by sorting by IP string
	sort.Slice(out, func(i, j int) bool {
		return out[i].String() < out[j].String()
	})
	return
}

func joinIPs(xs []ip.IP) string {
	if len(xs) == 0 {
		return ""
	}
	s := make([]string, len(xs))
	for i, a := range xs {
		s[i] = a.String()
	}
	return strings.Join(s, ", ")
}

func diff(a, b []ip.IP) ([]ip.IP, []ip.IP) {
	var onlyInA, onlyInB []ip.IP
	i, j := 0, 0

	for i < len(a) && j < len(b) {
		ka := a[i].String()
		kb := b[j].String()
		switch {
		case ka == kb:
			i++
			j++
		case ka < kb:
			onlyInA = append(onlyInA, a[i])
			i++
		default: // ka > kb
			onlyInB = append(onlyInB, b[j])
			j++
		}
	}

	for i < len(a) {
		onlyInA = append(onlyInA, a[i])
		i++
	}

	for j < len(b) {
		onlyInB = append(onlyInB, b[j])
		j++
	}

	return onlyInA, onlyInB
}
