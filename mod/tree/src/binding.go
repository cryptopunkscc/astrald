package tree

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

var _ tree.Binding = &binding{}

type binding struct {
	mod      *Module
	path     string
	node     tree.Node
	cancel   func()
	onChange func(astral.Object)

	mu        sync.RWMutex
	value     astral.Object
	closeOnce sync.Once
}

func (mod *Module) Bind(ctx *astral.Context, path string, defaultValue astral.Object, onChange func(astral.Object)) (tree.Binding, error) {
	node, err := tree.Query(ctx, mod.Root(), path, true)
	if err != nil {
		return nil, err
	}

	// Set default synchronously if needed (no lock needed yet)
	if defaultValue != nil {
		if err := node.Set(ctx, defaultValue); err != nil {
			return nil, err
		}
	}

	subCtx, cancel := ctx.WithCancel()
	ch, err := node.Get(subCtx, true)
	if err != nil {
		cancel()
		return nil, err
	}

	b := &binding{
		mod:      mod,
		path:     path,
		node:     node,
		cancel:   cancel,
		onChange: onChange,
	}

	go func() {
		for obj := range ch {
			b.mu.Lock()
			b.value = obj
			b.mu.Unlock()

			if b.onChange != nil {
				b.onChange(obj)
			}
		}
	}()

	mod.registerBinding(path, b)

	return b, nil
}

func (b *binding) Value() astral.Object {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.value
}

func (b *binding) Set(ctx *astral.Context, v astral.Object) error {
	return b.node.Set(ctx, v)
}

func (b *binding) Close() {
	b.closeOnce.Do(func() {
		b.cancel()
		b.mod.unregisterBinding(b.path, b)
	})
}
