package modules

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/config"
	"strings"
	"sync"
)

var _ Modules = &CoreModules{}

type CoreModules struct {
	loaded      map[string]Module
	enabled     []string
	node        Node
	configStore config.Store
	log         *log.Logger
}

func NewCoreModules(node Node, mods []string, configStore config.Store, log *log.Logger) (*CoreModules, error) {
	m := &CoreModules{
		log:         log.Tag("modules"),
		configStore: configStore,
		loaded:      make(map[string]Module),
		node:        node,
		enabled:     mods,
	}

	return m, nil
}

func (m *CoreModules) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	for _, name := range m.enabled {
		if err := m.Load(name); err != nil {
			m.log.Log("error loading module %s: %s", name, err)
			continue
		}
	}

	// log loaded module names
	var modNames = make([]string, 0, len(m.loaded))
	for name, _ := range m.loaded {
		modNames = append(modNames, name)
	}

	m.log.Log("enabled: %s", strings.Join(modNames, " "))

	for name, mod := range m.loaded {
		name, mod := name, mod
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := mod.Run(ctx)
			if err != nil {
				m.log.Error("module %s ended with error: %s",
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

func (m *CoreModules) Load(name string) error {
	loader, found := moduleLoaders[name]
	if !found {
		return errors.New("module not found")
	}

	mod, err := loader.Load(m.node, config.NewPrefixStore(m.configStore, "mod_"))
	if err != nil {
		return err
	}

	m.loaded[name] = mod

	return nil
}

func (m *CoreModules) Find(name string) Module {
	return m.loaded[name]
}

func (m *CoreModules) Loaded() []Module {
	var mods = make([]Module, 0, len(m.loaded))
	for _, mod := range m.loaded {
		mods = append(mods, mod)
	}
	return mods
}
