# Node

`astral.Node` is `Router + Identity()`. `core.Node` implements it and embeds
`*core.Router` (a `PriorityRouter` plus query preprocessors and a session
map). It owns the `Modules` manager and node assets.

Encrypted links to peers are provided by `mod/nodes`, not the core node.

## Module Lifecycle

Load -> Inject -> LoadDependencies -> Prepare -> Run

* Load: instantiate the module. Do not access other modules.
* Inject: `core` registers each module that implements `astral.Router` or
  `QueryPreprocessor` with the node.
* LoadDependencies: fill Deps structs with `core.Inject`; register resolvers
  and filters.
* Prepare: apply pre-run config. All dependencies are available.
* Run: block the goroutine until context cancellation.

## Scheduler

`mod/scheduler` runs tasks after all declared dependencies signal Done. It is
a module, not part of `core.Node`.
