# Node

`core.Node` combines:

* `PriorityRouter`
* query preprocessors
* module manager

It establishes encrypted Links to peers.

**Scheduler**

`mod/scheduler` runs tasks after all declared dependencies signal Done.

**Module Lifecycle**

Load -> Inject -> LoadDependencies -> Prepare -> Run

* Load: instantiate the module. Do not access other modules.
* Inject: register Router and Preprocessor modules with the node.
* LoadDependencies: fill Deps structs with `core.Inject`; register resolvers
  and filters.
* Prepare: apply pre-run config. All dependencies are available.
* Run: block the goroutine until context cancellation.
