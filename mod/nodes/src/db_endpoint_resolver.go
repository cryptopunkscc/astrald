package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

type DBEndpointResolver struct {
	mod *Module
}

// ResolveEndpoints is a nodes.EndpointResolver that fetches endpoints from the local database
func (r *DBEndpointResolver) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (<-chan exonet.Endpoint, error) {
	var ch = make(chan exonet.Endpoint)

	rows, err := r.mod.db.FindEndpoints(nodeID)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)
		
		for _, row := range rows {
			// parse the next row
			e, err := r.mod.Exonet.Parse(row.Network, row.Address)
			if err != nil {
				r.mod.log.Errorv(1, "DB.ResolveEndpoints: error parsing db row: %v", err)
				continue
			}

			// send the next endpoint unless the context has been canceled
			select {
			case ch <- e:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

func (r *DBEndpointResolver) String() string { return "DBEndpointResolver" }
