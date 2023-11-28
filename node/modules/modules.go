package modules

import (
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
