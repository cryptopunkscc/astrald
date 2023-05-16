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
	var loaded = make([]string, 0, len(m.enabled))

	for _, name := range m.enabled {
		if err := m.Load(name); err != nil {
			m.log.Error("load %s: %s", name, err)
		} else {
			loaded = append(loaded, name)
		}
	}

	var started = make([]string, 0, len(loaded))
	for _, name := range loaded {
		mod, ok := m.loaded[name]
		if !ok {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := mod.Run(ctx)
			if err != nil {
				m.log.Error("run %s: %s", name, err)
			}
		}()
		started = append(started, name)
	}

	m.log.Log("started: %s", strings.Join(started, " "))

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
