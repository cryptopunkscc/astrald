package nodes

import (
	"slices"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
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

func (mod *Module) ResolveEndpointsFiltered(ctx *astral.Context, target *astral.Identity, o *RetrieveLinkOptions) (_ <-chan exonet.Endpoint, err error) {
	resolve, err := mod.ResolveEndpoints(ctx, target)
	if err != nil {
		return nil, err
	}

	if len(o.IncludeNetworks) == 0 && len(o.ExcludeNetworks) == 0 {
		return resolve, nil
	}

	ch := make(chan exonet.Endpoint)
	go func() {
		defer close(ch)
		for e := range resolve {
			net := e.Network()
			if len(o.ExcludeNetworks) > 0 && slices.Contains(o.ExcludeNetworks, net) {
				continue
			}
			if len(o.IncludeNetworks) > 0 && !slices.Contains(o.IncludeNetworks, net) {
				continue
			}

			select {
			case ch <- e:
			case <-ctx.Done():
				return
			}
		}
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
