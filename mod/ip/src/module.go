package ip

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ ip.Module = &Module{}

type Deps struct {
	Objects objects.Module
}

type Module struct {
	Deps
	node   astral.Node
	log    *log.Logger
	assets resources.Resources
	ops    shell.Scope

	providers sig.Set[ip.PublicIPCandidateProvider]
}

func (mod *Module) Run(ctx *astral.Context) error {
	go mod.watchAddresses(ctx)

	<-ctx.Done()
	return nil
}

func (mod *Module) LocalIPs() ([]ip.IP, error) {
	return mod.localAddresses(false)
}

func (mod *Module) localAddresses(includeLoopback bool) (out []ip.IP, err error) {
	ifaceAddrs, err := InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range ifaceAddrs {
		ipnet, ok := addr.(*net.IPNet)
		switch {
		case !ok:
			continue
		case ipnet.IP == nil:
			continue
		case ipnet.IP.IsLoopback() && !includeLoopback:
			continue
		}

		out = append(out, ip.IP(ipnet.IP))
	}

	return
}

func (mod *Module) watchAddresses(ctx context.Context) {
	addrs, err := mod.localAddresses(false)
	if err != nil {
		mod.log.Errorv(0,
			"network interface monitoring disabled, because fetching addresses failed: %v", err)
		return
	}

	for {
		select {
		case <-time.After(3 * time.Second):
			newAddrs, err := mod.localAddresses(false)
			if err != nil {
				mod.log.Errorv(0,
					"get network addresses: %v",
					err)
			}

			removed, added := sig.SliceDiffFunc(addrs, newAddrs, func(a, b ip.IP) int {
				return strings.Compare(a.String(), b.String())
			})

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

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) String() string {
	return ip.ModuleName
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
