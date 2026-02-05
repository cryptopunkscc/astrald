package nodes

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

type NodeLinker struct {
	mod        *Module
	Target     *astral.Identity
	produced   chan *Stream
	strategies map[string]LinkStrategy
}

func NewNodeLinker(mod *Module, target *astral.Identity) *NodeLinker {
	return &NodeLinker{
		mod:        mod,
		Target:     target,
		produced:   make(chan *Stream, 1),
		strategies: make(map[string]LinkStrategy),
	}
}

func (linker *NodeLinker) Produced() <-chan *Stream {
	return linker.produced
}

func (linker *NodeLinker) AddStrategy(network string, strategy LinkStrategy) {
	linker.strategies[network] = strategy
}

func (linker *NodeLinker) Activate(ctx *astral.Context, constraints LinkConstraints) {
	var allEndpoints []exonet.Endpoint

	if len(constraints.Endpoints) > 0 {
		allEndpoints = constraints.Endpoints
	} else {
		resolved, err := linker.mod.ResolveEndpoints(ctx, linker.Target)
		if err != nil {
			return
		}
		for ep := range resolved {
			allEndpoints = append(allEndpoints, ep)
		}
	}

	filter := endpointFilter(constraints.IncludeNetworks, constraints.ExcludeNetworks)

	for net, strategy := range linker.strategies {
		var eps = make([]exonet.Endpoint, 0, len(allEndpoints))
		for _, ep := range allEndpoints {
			if !filter(ep) {
				continue
			}

			if ep.Network() != net {
				continue
			}
			eps = append(eps, ep)
		}

		if len(eps) == 0 {
			continue
		}

		strategy.Activate(ctx, eps)
	}
}

type LinkConstraints struct {
	IncludeNetworks []string
	ExcludeNetworks []string
	Endpoints       []exonet.Endpoint
}

func endpointFilter(include, exclude []string) func(exonet.Endpoint) bool {
	return func(endpoint exonet.Endpoint) bool {
		net := endpoint.Network()

		if len(exclude) > 0 && slices.Contains(exclude, net) {
			return false
		}

		if len(include) > 0 && !slices.Contains(include, net) {
			return false
		}

		return true
	}
}

type LinkStrategy interface {
	Activate(ctx *astral.Context, endpoints []exonet.Endpoint) error
}
