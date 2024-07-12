package core

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
)

type ModuleLoader interface {
	Load(node.Node, assets.Assets, *log.Logger) (node.Module, error)
}

type DependencyLoader interface {
	LoadDependencies() error
}

type Preparer interface {
	Prepare(context.Context) error
}

var moduleLoaders = map[string]ModuleLoader{}

func RegisterModule(name string, loader ModuleLoader) error {
	if _, found := moduleLoaders[name]; found {
		return errors.New("module already added")
	}

	moduleLoaders[name] = loader

	return nil
}

func RegisteredModules() []string {
	var list = make([]string, 0, len(moduleLoaders))
	for m := range moduleLoaders {
		list = append(list, m)
	}
	return list
}
