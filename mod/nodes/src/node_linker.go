package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
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

// todo: will probably pass some kind of options in the future
func (linker *NodeLinker) Activate(ctx *astral.Context) {
	for _, strategy := range linker.strategies {
		strategy.Activate(ctx)
	}
}

type LinkStrategy interface {
	// todo: will probably accept some kind of options in the future
	Activate(ctx *astral.Context) error
}
