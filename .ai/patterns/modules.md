# Module Patterns

Use for module structure, lifecycle, or dependency injection changes.

## Module Layout

```text
mod/<name>/
  module.go      public interface (exported types, constants, method name strings)
  src/
    loader.go    Loader{}, Load(), init() registration
    module.go    Module struct, Run()
    deps.go      Deps struct, LoadDependencies()
    config.go    Config struct, defaults
    op_<name>.go one file per op handler
```

## Loader Registration

```go
type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
    mod := &Module{node: node, log: log}
    _ = assets.LoadYAML(mymodule.ModuleName, &mod.config)
    mod.router.AddStructPrefix(mod, "Op")
    return mod, nil
}

func init() {
    if err := core.RegisterModule(mymodule.ModuleName, Loader{}); err != nil {
        panic(err)
    }
}
```

Source: `mod/nodes/src/loader.go`

## Dependencies

Inject dependencies in `LoadDependencies`, not `Load`.

```go
type Deps struct {
    Auth    auth.Module
    Objects objects.Module
    Dir     dir.Module
}

func (mod *Module) LoadDependencies(*astral.Context) error {
    return core.Inject(mod.node, &mod.Deps)
}
```

Use a struct tag to override a registry name: `` Dir dir.Module `mod:"directory"` ``.
`core.Inject` fails if any required module is absent.

Source: `mod/nodes/src/deps.go`

## Run

```go
func (mod *Module) Run(ctx *astral.Context) error {
    mod.ctx = ctx.IncludeZone(astral.ZoneNetwork) // only if module needs network
    <-mod.Deps.Scheduler.Ready()                  // only if using scheduler
    <-ctx.Done()
    return nil
}
```

Source: `mod/nodes/src/module.go`
