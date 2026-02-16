package nodes

import (
	"fmt"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	nodescli "github.com/cryptopunkscc/astrald/mod/nodes/client"
)

func (mod *Module) ResolveEndpoints(ctx *astral.Context, nodeID *astral.Identity) (_ <-chan *nodes.ResolvedEndpoint, err error) {
	var ch = make(chan *nodes.ResolvedEndpoint)

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

func (mod *Module) runResolver(ctx *astral.Context, r nodes.EndpointResolver, nodeID *astral.Identity, out chan<- *nodes.ResolvedEndpoint) {
	ch, err := r.ResolveEndpoints(ctx, nodeID)
	if err != nil {
		return
	}
	for {
		var e *nodes.ResolvedEndpoint
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

func (mod *Module) updateNodeEndpoints(ctx *astral.Context, identity *astral.Identity) error {
	client := nodescli.New(identity, astrald.Default())

	endpoints, err := client.ResolveEndpoints(ctx.IncludeZone(astral.ZoneNetwork), identity)
	if err != nil {
		return fmt.Errorf("resolve endpoints: %v", err)
	}

	for _, ep := range endpoints {
		err = mod.AddEndpoint(identity, ep)
		if err != nil {
			mod.log.Error("adding resolved endpoint failed %v: %v", ep, err)
		}
	}

	return nil
}
