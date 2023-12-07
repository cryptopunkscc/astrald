package modules

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
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
	var wg sync.WaitGroup

	// Step 1 - load all modules
	var loaded = make([]string, 0, len(m.enabled))
	for _, name := range m.enabled {
		if err := m.Load(name); err != nil {
			m.log.Error("load %s: %s", name, err)
		} else {
			loaded = append(loaded, name)
		}
	}

	// Step 2 - configure modules
	var prepared = make([]string, 0, len(loaded))
	for _, name := range loaded {
		if p, ok := m.loaded[name].(Preparer); ok {
			err := p.Prepare(ctx)
			if err != nil {
				m.log.Error("module %s error: %v", name, err)
				continue
			}
		}
		prepared = append(prepared, name)
	}

	// Step 3 - run modules
	var started = make([]string, 0, len(prepared))
	for _, name := range prepared {
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

func (m *CoreModules) Load(name string) error {
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
