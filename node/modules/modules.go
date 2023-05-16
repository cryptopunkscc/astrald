package modules

import (
	"context"
	"errors"
)

type Modules interface {
	Find(name string) Module
	Loaded() []Module
}

func Find[T Module](modules Modules) (mod T, err error) {
	for _, m := range modules.Loaded() {
		var ok bool
		if mod, ok = m.(T); ok {
			return
		}
	}

	return mod, errors.New("module not found")
}

type ReadyWaiter interface {
	Ready() <-chan struct{}
}

func WaitReady[T ReadyWaiter](ctx context.Context, modules Modules) (mod T, err error) {
	for _, m := range modules.Loaded() {
		var ok bool
		if mod, ok = m.(T); ok {
			select {
			case <-ctx.Done():
				return mod, ctx.Err()
			case <-mod.Ready():
				return mod, nil
			}
		}
	}

	return mod, errors.New("module not found")
}
