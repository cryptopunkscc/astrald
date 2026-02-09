package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

type NodeLinker struct {
	mod        *Module
	Target     *astral.Identity
	strategies map[string]nodes.LinkStrategy
}

func NewNodeLinker(mod *Module, target *astral.Identity) *NodeLinker {
	linker := &NodeLinker{
		mod:        mod,
		Target:     target,
		strategies: make(map[string]nodes.LinkStrategy),
	}

	for network, factory := range mod.strategyFactories.Clone() {
		linker.strategies[network] = factory.Build(target)
	}

	return linker
}

func (linker *NodeLinker) Activate(ctx *astral.Context, networks []string) <-chan struct{} {
	done := make(chan struct{})

	var strategies []nodes.LinkStrategy
	if len(networks) == 0 {
		for _, strategy := range linker.strategies {
			strategies = append(strategies, strategy)
		}
	} else {
		for _, network := range networks {
			if strategy, ok := linker.strategies[network]; ok {
				strategies = append(strategies, strategy)
			}
		}
	}

	if len(strategies) == 0 {
		close(done)
		return done
	}

	var doneChannels []<-chan struct{}
	for _, strategy := range strategies {
		doneChannels = append(doneChannels, strategy.Signal(ctx))
	}

	return sig.WaitAllDone(ctx, doneChannels...)
}
