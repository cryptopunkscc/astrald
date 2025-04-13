package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

func (mod *Module) ResolveEndpoints(ctx context.Context, nodeID *astral.Identity) (endpoints []exonet.Endpoint, err error) {
	var rows []dbEndpoint

	err = mod.db.Find(&rows, "identity = ?", nodeID).Error
	if err != nil {
		return
	}

	for _, row := range rows {
		e, err := mod.Exonet.Parse(row.Network, row.Address)
		if err != nil {
			mod.log.Errorv(1, "Endpoints(): error parsing db row: %v", err)
			continue
		}
		endpoints = append(endpoints, e)
	}

	return
}
