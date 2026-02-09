package core

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	log2 "github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/sig"
)

type Modules struct {
	loaded  map[string]Module
	enabled []string
	node    *Node
	assets  assets.Assets
	log     *log2.Logger
}

type Module interface {
	Run(*astral.Context) error
}

type ModuleLoader interface {
	Load(astral.Node, assets.Assets, *log2.Logger) (Module, error)
}

var modules = sig.Map[string, ModuleLoader]{}

func NewModules(n *Node, mods []string, assets assets.Assets, log *log2.Logger) (*Modules, error) {
	m := &Modules{
		log:     log.Tag("modules"),
		assets:  assets,
		loaded:  make(map[string]Module),
		node:    n,
		enabled: mods,
	}

	return m, nil
}

func (m *Modules) Run(ctx *astral.Context) error {
	// Load enabled modules. Loaders should only return a new instance of the module and must not try
	// to access other modules, as the order of loading is undefined.
	var loaded = m.loadEnabled()

	m.injectLoaded()

	// Dependencies - in this stage modules load their deps and injects its handlers.
	var deps = m.loadDependecies(ctx, loaded)

	// Prepare - In this stage, modules should perform any configuration necessary before services are run.
	var prepared = m.prepareModules(ctx, deps)

	// Run modules. During this stage modules should run all their services for the duration of the context.
	return m.runModules(ctx, prepared)
}

func (m *Modules) loadEnabled() []string {
	// Sage 1 - Load. During this stage
	var loaded = make([]string, 0, len(m.enabled))
	for _, name := range m.enabled {
		if err := m.loadModule(name); err != nil {
			m.log.Error("load %v: %v", name, err)
		} else {
			loaded = append(loaded, name)
		}
	}
	return loaded
}

func (m *Modules) injectLoaded() {
	var routers []any
	for _, mod := range m.loaded {
		if p, ok := mod.(QueryPreprocessor); ok {
			m.node.AddQueryPreprocessor(p)
		}

		if r, ok := mod.(astral.Router); ok {
			m.node.Add(r, 0)
			routers = append(routers, r)
		}
	}

	if len(routers) > 0 {
		m.log.Logv(2, "routers: %v"+strings.Repeat(", %v", len(routers)-1), routers...)
	}
}

func (m *Modules) loadDependecies(ctx *astral.Context, modules []string) []string {
	var loaded sig.Set[string]

	var wg sync.WaitGroup
	for _, name := range modules {
		name := name

		wg.Add(1)
		go func() {
			defer wg.Done()
			if p, ok := m.loaded[name].(DependencyLoader); ok {
				err := p.LoadDependencies(ctx)
				if err != nil {
					m.log.Error("module %v load dependencies: %v", name, err)
					panic(err) // TODO: handle this cleanly instead of panicking
					return
				}
			}
			loaded.Add(name)
		}()
	}

	wg.Wait()

	return loaded.Clone()
}

func (m *Modules) prepareModules(ctx context.Context, modules []string) []string {
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
					m.log.Error("module %v prepare: %v", name, err)
					return
				}
			}
			prepared.Add(name)
		}()
	}

	wg.Wait()

	return prepared.Clone()
}

func (m *Modules) runModules(ctx *astral.Context, modules []string) error {
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
				m.log.Error("module %v panicked: %v", name, p)
			})

			defer wg.Done()

			err := mod.Run(ctx)
			switch {
			case err == nil:
			case errors.Is(err, context.Canceled):
			default:
				m.log.Error("module %v ended with error: %v", name, err)
			}
		}()
		started = append(started, name)
	}

	slices.Sort(started)

	m.log.Log("started: %v", strings.Join(started, " "))

	// wait for all modules to finish
	wg.Wait()

	return nil
}

func (m *Modules) loadModule(name string) error {
	loader, found := modules.Get(name)
	if !found {
		return errors.New("module not found")
	}

	mod, err := loader.Load(m.node, m.assets, m.log.Tag(log2.Tag(name)))
	if err != nil {
		return err
	}

	m.loaded[name] = mod

	return nil
}

func (m *Modules) Find(name string) Module {
	return m.loaded[name]
}

func (m *Modules) Loaded() []Module {
	var mods = make([]Module, 0, len(m.loaded))
	for _, mod := range m.loaded {
		mods = append(mods, mod)
	}
	return mods
}

func RegisterModule(name string, loader ModuleLoader) error {
	if _, ok := modules.Set(name, loader); !ok {
		return errors.New("module already added")
	}

	return nil
}

func Load[M any](node astral.Node, name string) (M, error) {
	cnode, ok := node.(*Node)
	if !ok {
		var m M
		return m, errors.New("unsupported node type")
	}
	mod, ok := cnode.Modules().Find(name).(M)
	if !ok {
		return mod, errModuleUnavailable(name)
	}
	return mod, nil
}

// Inject injects modules into a struct
func Inject(node astral.Node, target any) (err error) {
	cnode, ok := node.(*Node)
	if !ok {
		return errors.New("unsupported node type")
	}

	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		modName := strings.ToLower(fieldType.Name)
		if tag := fieldType.Tag.Get("mod"); tag != "" {
			modName = tag
		}

		if field.CanSet() {
			mod := cnode.Modules().Find(modName)
			if mod == nil {
				return fmt.Errorf("cannot find module %s", modName)
			}

			modVal := reflect.ValueOf(mod)
			if modVal.Type().AssignableTo(field.Type()) {
				field.Set(modVal)
			} else {
				return fmt.Errorf("cannot inject field %s", fieldType.Name)
			}
		}
	}
	return
}

type DependencyLoader interface {
	LoadDependencies(ctx *astral.Context) error
}

type Preparer interface {
	Prepare(context.Context) error
}

func EachLoadedModule(node astral.Node, fn func(Module) error) (err error) {
	coreNode, ok := node.(*Node)
	if !ok {
		return errors.New("unsupported node type")
	}
	for _, m := range coreNode.Modules().Loaded() {
		err = fn(m)
		if err != nil {
			return
		}
	}
	return
}
