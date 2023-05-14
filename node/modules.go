package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/config"
	"strings"
	"sync"
)

type ModuleLoader interface {
	Load(node Node, configStore config.Store) (Module, error)
	Name() string
}

type Module interface {
	Run(context.Context) error
}

type ModuleManager struct {
	modules map[string]Module
	node    Node
}

func NewModuleManager(node Node, loaders []ModuleLoader, configStore config.Store) (*ModuleManager, error) {
	m := &ModuleManager{
		modules: make(map[string]Module),
		node:    node,
	}

	prefixStore := config.NewPrefixStore(configStore, "mod_")

	for _, loader := range loaders {
		name := loader.Name()
		mod, err := loader.Load(node, prefixStore)
		if err != nil {
			log.Log("error loading module %s: %s", name, err)
			continue
		}
		m.modules[name] = mod
	}

	return m, nil
}

func (manager *ModuleManager) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	// log loaded module names
	var modNames = make([]string, 0, len(manager.modules))
	for name, _ := range manager.modules {
		modNames = append(modNames, name)
	}

	log.Log("modules: %s", strings.Join(modNames, " "))

	for name, mod := range manager.modules {
		name, mod := name, mod
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := mod.Run(ctx)
			if err != nil {
				log.Error("module %s ended with error: %s",
					name,
					err,
				)
			}
		}()
	}

	// wait for all modules to finish
	wg.Wait()
	return nil
}

func (manager *ModuleManager) FindModule(name string) Module {
	return manager.modules[name]
}
