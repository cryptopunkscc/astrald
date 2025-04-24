package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"sync"
)

func (mod *Module) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (_ <-chan exonet.Endpoint, err error) {
	var ch = make(chan exonet.Endpoint)

	go func() {
		var wg sync.WaitGroup

		// spawn all resolvers
		for _, r := range mod.resolvers.Clone() {
			r := r
			wg.Add(1)
			go func() {
				defer wg.Done()
				mod.runResolver(ctx, r, nodeID, ch)
			}()
		}

		// wait for all resolvers to finish
		wg.Wait()
		close(ch)
	}()

	return ch, nil
}

func (mod *Module) runResolver(ctx *astral.Context, r nodes.EndpointResolver, nodeID *astral.Identity, out chan<- exonet.Endpoint) {
	ch, err := r.ResolveEndpoints(ctx, nodeID)
	if err != nil {
		return
	}
	for {
		var e exonet.Endpoint
		var ok bool

		// read the next endpoint from the resolver
		select {
		case e, ok = <-ch:
			if !ok {
				return
			}
		case <-ctx.Done():
			return
		}

		// write the endpoint upstream
		select {
		case out <- e:
		case <-ctx.Done():
			return
		}
	}
}
