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
func (linker *NodeLinker) Activate(ctx *astral.Context) chan error {
	done := make(chan error, 1)

	var strategiesDone []chan error

	for _, strategy := range linker.strategies {
		strategyDone := strategy.Activate(ctx)
		strategiesDone = append(strategiesDone, strategyDone)
	}

	// note: ensures that every strategy is done
	go func() {
		for _, ch := range strategiesDone {
			select {
			case <-ctx.Done():
				done <- ctx.Err()
				return
			case <-ch:
			}
		}
		done <- nil
	}()

	return done
}

type LinkStrategy interface {
	// todo: will probably accept some kind of options in the future
	Activate(ctx *astral.Context) chan error
}
