package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"sync"
)

func (mod *Module) Search(ctx context.Context, query string, opts *objects.SearchOpts) (<-chan *objects.SearchResult, error) {
	if !opts.Zone.Is(astral.ZoneNetwork) {
		return nil, astral.ErrZoneExcluded
	}

	var results = make(chan *objects.SearchResult)

	go func() {
		defer close(results)

		v, ok := opts.Extra.Get("ext")
		if !ok {
			return
		}

		nodes, ok := v.([]*astral.Identity)
		if !ok || len(nodes) == 0 {
			return
		}

		var wg sync.WaitGroup
		for _, nodeID := range nodes {
			nodeID := nodeID

			if opts.QueryFilter != nil {
				if !opts.QueryFilter(nodeID) {
					continue
				}
			}

			wg.Add(1)
			go func() {
				defer wg.Done()

				c, err := mod.Objects.Connect(nodeID, opts.ClientID)
				if err != nil {
					mod.log.Errorv(1, "objects.connect %v: %v", nodeID, err)
					return
				}

				_results, err := c.Search(ctx, query)
				if err != nil {
					mod.log.Errorv(1, "objects.search %v: %v", nodeID, err)
					return
				}

				for r := range _results {
					mod.searchCache.Set(r.ObjectID.String(), nodeID)
					results <- r
				}
			}()
		}
		wg.Wait()
	}()

	return results, nil
}
