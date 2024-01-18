package modules

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/sig"
	"slices"
	"strings"
	"sync"
)

var _ Modules = &CoreModules{}

type CoreModules struct {
	loaded  map[string]Module
	enabled []string
	node    Node
	assets  assets.Store
	log     *log.Logger
}

func NewCoreModules(node Node, mods []string, assets assets.Store, log *log.Logger) (*CoreModules, error) {
	m := &CoreModules{
		log:     log.Tag("modules"),
		assets:  assets,
		loaded:  make(map[string]Module),
		node:    node,
		enabled: mods,
	}

	return m, nil
}

func (m *CoreModules) Run(ctx context.Context) error {
	// Load enabled modules. Loaders should only return a new instance of the module and must not try
	// to access other modules, as the order of loading is undefined.
	var loaded = m.loadEnabled()

	// Dependencies - in this stage modules load their deps and injects its handlers.
	var deps = m.loadDependecies(loaded)

	// Prepare - In this stage, modules should perform any configuration necessary before services are run.
	var prepared = m.prepareModules(ctx, deps)

	// Run modules. During this stage modules should run all their services for the duration of the context.
	return m.runModules(ctx, prepared)
}

func (m *CoreModules) loadEnabled() []string {
	// Sage 1 - Load. During this stage
	var loaded = make([]string, 0, len(m.enabled))
	for _, name := range m.enabled {
		if err := m.loadModule(name); err != nil {
			m.log.Error("load %s: %s", name, err)
		} else {
			loaded = append(loaded, name)
		}
	}
	return loaded
}

func (m *CoreModules) loadDependecies(modules []string) []string {
	var loaded sig.Set[string]

	var wg sync.WaitGroup
	for _, name := range modules {
		name := name

		wg.Add(1)
		go func() {
			defer wg.Done()
			if p, ok := m.loaded[name].(DependencyLoader); ok {
				err := p.LoadDependencies()
				if err != nil {
					m.log.Error("module %s load dependencies: %v", name, err)
					return
				}
			}
			loaded.Add(name)
		}()
	}

	wg.Wait()

	return loaded.Clone()
}

func (m *CoreModules) prepareModules(ctx context.Context, modules []string) []string {
	var prepared sig.Set[string]

	var wg sync.WaitGroup
	for _, name := range modules {
		name := name

		wg.Add(1)
		go func() {
			defer wg.Done()
			if p, ok := m.loaded[name].(Preparer); ok {
				err := p.Prepare(ctx)
				if err != nil {
					m.log.Error("module %s prepare: %v", name, err)
					return
				}
			}
			prepared.Add(name)
		}()
	}

	wg.Wait()

	return prepared.Clone()
}

func (m *CoreModules) runModules(ctx context.Context, modules []string) error {
	var wg sync.WaitGroup

	var started = make([]string, 0, len(modules))
	for _, name := range modules {
		mod, ok := m.loaded[name]
		if !ok {
			continue
		}

		name := name
		wg.Add(1)
		go func() {
			defer debug.SaveLog(func(p any) {
				m.log.Error("module %s panicked: %v", name, p)
			})

			defer wg.Done()

			err := mod.Run(ctx)
			switch {
			case err == nil:
			case errors.Is(err, context.Canceled):
			default:
				m.log.Error("module %s ended with error: %s", name, err)
			}
		}()
		started = append(started, name)
	}

	slices.Sort(started)

	m.log.Log("started: %s", strings.Join(started, " "))

	// wait for all modules to finish
	wg.Wait()

	return nil
}

func (m *CoreModules) loadModule(name string) error {
	loader, found := moduleLoaders[name]
	if !found {
		return errors.New("module not found")
	}

	mod, err := loader.Load(m.node, assets.NewPrefixStore(m.assets, "mod_"), m.log.Tag(name))
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
