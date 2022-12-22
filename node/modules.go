package node

import (
	"context"
	"log"
	"strings"
	"sync"
)

type ModuleLoader interface {
	Load(node *Node) (Module, error)
	Name() string
}

type Module interface {
	Run(context.Context) error
}

type ModuleManager struct {
	modules map[string]Module
}

func NewModuleManager(node *Node, loaders []ModuleLoader) (*ModuleManager, error) {
	m := &ModuleManager{
		modules: make(map[string]Module),
	}

	for _, loader := range loaders {
		name := loader.Name()
		mod, err := loader.Load(node)
		if err != nil {
			log.Printf("error loading module %s: %s", name, err.Error())
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
	log.Println("running modules:", strings.Join(modNames, " "))

	for name, mod := range manager.modules {
		name, mod := name, mod
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := mod.Run(ctx)
			if err != nil {
				log.Printf("module %s ended with error: %s", name, err.Error())
			}
		}()
	}

	// wait for all modules to finish
	wg.Wait()
	return nil
}
