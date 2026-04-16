package nearby

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ nodes.EndpointResolver = &Module{}

func (mod *Module) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (_ <-chan *nodes.EndpointWithTTL, err error) {
	var list []*nodes.EndpointWithTTL

	for _, v := range mod.Cache().Clone() {
		if v.GetIdentity() == nil {
			continue
		}

		if !v.GetIdentity().IsEqual(nodeID) {
			continue
		}

		endpoints := astral.SelectByType[*nodes.EndpointWithTTL](v.Status.Attachments.Objects())
		if len(endpoints) > 0 {
			list = append(list, endpoints...)
			continue
		}
	}

	return sig.ArrayToChan(list), nil
}
