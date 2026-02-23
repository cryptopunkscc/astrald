package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

type NodeLinker struct {
	mod            *Module
	Target         *astral.Identity
	strategies     sig.Map[string, nodes.LinkStrategy]
	autoStrategies sig.Set[string]
}

func NewNodeLinker(mod *Module, target *astral.Identity) *NodeLinker {
	linker := &NodeLinker{
		mod:    mod,
		Target: target,
	}

	for network, factory := range mod.strategyFactories.Clone() {
		linker.strategies.Set(network, factory.Build(target))
	}

	linker.autoStrategies.Add(mod.autoStrategies.Clone()...)

	return linker
}

func (linker *NodeLinker) Activate(ctx *astral.Context, strategies []string) <-chan struct{} {
	if len(strategies) == 0 {
		strategies = linker.autoStrategies.Clone()
	}

	var doneChannels []<-chan struct{}
	for _, network := range strategies {
		if strategy, ok := linker.strategies.Get(network); ok {
			strategy.Signal(ctx)
			doneChannels = append(doneChannels, strategy.Done())
		}
	}

	return sig.WaitAllDone(ctx, doneChannels...)
}
