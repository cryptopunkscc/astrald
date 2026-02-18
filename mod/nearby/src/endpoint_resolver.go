package nearby

import (
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ nodes.EndpointResolver = &Module{}

func (mod *Module) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (_ <-chan *nodes.EndpointWithTTL, err error) {
	var list []*nodes.EndpointWithTTL

	for _, v := range mod.Cache().Clone() {
		if !v.Identity.IsEqual(nodeID) {
			continue
		}

		hostport := fmt.Sprintf("%s:%d", v.IP, v.Status.Port)

		ep, err := tcp.ParseEndpoint(hostport)
		if err != nil {
			continue
		}

		list = append(list, nodes.NewEndpointWithTTL(ep, 24*time.Hour))
	}

	return sig.ArrayToChan(list), nil
}
