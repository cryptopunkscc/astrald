package status

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

func (mod *Module) ResolveEndpoints(ctx context.Context, identity *astral.Identity) (list []exonet.Endpoint, err error) {
	for _, v := range mod.Cache().Clone() {
		if !v.Identity.IsEqual(identity) {
			continue
		}

		hostport := fmt.Sprintf("%s:%d", v.IP, v.Status.Port)

		ep, err := tcp.ParseEndpoint(hostport)
		if err != nil {
			continue
		}

		list = append(list, ep)
	}

	return
}
