package modules

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/node/assets"
)

type ModuleLoader interface {
	Load(Node, assets.Store) (Module, error)
}

type Module interface {
	Run(context.Context) error
}

var moduleLoaders = map[string]ModuleLoader{}

func RegisterModule(name string, loader ModuleLoader) error {
	if _, found := moduleLoaders[name]; found {
		return errors.New("module already added")
	}

	moduleLoaders[name] = loader

	return nil
}
