package status

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/sig"
)

func (mod *Module) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (_ <-chan exonet.Endpoint, err error) {
	var list []exonet.Endpoint

	for _, v := range mod.Cache().Clone() {
		if !v.Identity.IsEqual(nodeID) {
			continue
		}

		hostport := fmt.Sprintf("%s:%d", v.IP, v.Status.Port)

		ep, err := tcp.ParseEndpoint(hostport)
		if err != nil {
			continue
		}

		list = append(list, ep)
	}

	return sig.ArrayToChan(list), nil
}
