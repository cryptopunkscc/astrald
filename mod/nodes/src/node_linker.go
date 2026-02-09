package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type NodeLinker struct {
	mod    *Module
	Target *astral.Identity
}

func NewNodeLinker(mod *Module, target *astral.Identity) *NodeLinker {
	return &NodeLinker{
		mod:    mod,
		Target: target,
	}
}

func (linker *NodeLinker) Activate(ctx *astral.Context, networks []string) <-chan struct{} {
	done := make(chan struct{})

	// get all registered networks if none specified
	if len(networks) == 0 {
		for network := range linker.mod.strategyFactories.Clone() {
			networks = append(networks, network)
		}
	}

	// build strategies for this activation
	var strategies []nodes.LinkStrategy
	for _, network := range networks {
		factory, ok := linker.mod.strategyFactories.Get(network)
		if !ok {
			continue
		}
		strategies = append(strategies, factory.Build(linker.Target))
	}

	if len(strategies) == 0 {
		close(done)
		return done
	}

	for _, strategy := range strategies {
		strategy.Signal(ctx)
	}

	go func() {
		for _, strategy := range strategies {
			select {
			case <-ctx.Done():
				close(done)
				return
			case <-strategy.Done():
			}
		}
		close(done)
	}()

	return done
}
