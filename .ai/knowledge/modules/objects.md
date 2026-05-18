# mod/objects

Hosts the node's content-addressed object layer behind a uniform query and repository surface. Owns default repository groups, object type indexing, object reads and writes, local receive and push dispatch, and extension registries for describers, searchers, search preprocessors, finders, holders, and receivers.

## Dependencies

| Module | Why |
|---|---|
| `auth` | `OpRead` authorizes `ReadObjectAction` before serving object bytes |
| `dir` | `Fetch` parses `astral://` ARLs with `arl.Parse` and resolves addresses through the directory |
| `nodes` (opt) | injected for module integration; network object calls are made through typed `objects/client` calls routed by the node |
| `core/assets` | loads `objects.yaml`, creates default resources, and backs the `objects__objects` database table |
| `core.Node` | `LoadDependencies` scans loaded modules for object extension interfaces |

## Flows

- Store object: `Store` -> repository `Create` -> canonical type encode into writer -> `Commit` -> return `*ObjectID`.
- Load object: repository `Read` -> read bytes -> canonical decode -> invalid magic bytes become `*astral.Blob`; other decode errors propagate.
- Create stream: `objects.create` accepts the query -> drains incoming `*astral.Blob` chunks into a repository writer -> commit message commits and returns `*ObjectID` -> deferred discard protects unfinished writers.
- Read stream: `objects.read` resolves repo and range -> `Authorize(ReadObjectAction)` -> repository `Read` -> streams bytes to the accepted channel.
- Repository group read: sequential groups walk members in priority order; concurrent groups race member reads and return the first successful reader.
- Repository group scan: each member scans in its own goroutine -> snapshot/follow boundary nils are collapsed into a single nil -> live updates continue while following.
- Search: preprocessors mutate `objects.Search` -> local searchers run concurrently -> network-zone searches query each requested source through `objects/client` -> operation handler deduplicates by object ID and applies optional repo filter.
- Receive local object: `Receive` normalizes zero source to the node identity -> builds a `Drop` backed by `WriteDefault` -> calls all receivers -> accepted drops return success; `Accept(true)` stores at most once.
- Push object: local target calls `Receive`; remote target calls `objects.push` through the typed client and expects a boolean result.
- Fetch object: HTTP and HTTPS fetch with `http.Get` then write to the default repo; `astral://` ARLs route an in-flight query and write the response to the default repo.
- Extension discovery: `LoadDependencies` walks loaded modules and auto-registers any object describer, searcher, search preprocessor, finder, holder, or receiver.
- Repository removal: `RemoveRepository` deletes the repo from the registry -> removes it from every group -> calls `AfterRemoved` when the repo implements the callback.

## Source

- `mod/objects/module.go`, `mod/objects/repository.go`, `mod/objects/writer.go`, `mod/objects/reader.go`, `mod/objects/read_seeker.go`, `mod/objects/load.go`, `mod/objects/errors.go` - public module, repository, reader, writer, and error contracts.
- `mod/objects/descriptor.go`, `mod/objects/search.go`, `mod/objects/search_query.go`, `mod/objects/search_result.go`, `mod/objects/find.go`, `mod/objects/probe.go`, `mod/objects/type_spec.go`, `mod/objects/read_object_action.go`, `mod/objects/create_object_action.go` - extension types, query data, probes, and auth actions.
- `mod/objects/src/loader.go`, `mod/objects/src/deps.go`, `mod/objects/src/module.go`, `mod/objects/src/config.go`, `mod/objects/src/db.go`, `mod/objects/src/db_object.go` - default repo layout, dependency injection, config, and type index persistence.
- `mod/objects/src/repo_group.go` - sequential and concurrent group behavior for read, create, contains, delete, scan, and free-space operations.
- `mod/objects/src/drop.go`, `mod/objects/src/receive.go`, `mod/objects/src/push.go`, `mod/objects/src/fetch.go`, `mod/objects/src/network_reader.go` - object receive, save-on-accept, push, fetch, and routed network reads.
- `mod/objects/src/describe.go`, `mod/objects/src/search.go`, `mod/objects/src/find.go`, `mod/objects/src/holding.go` - extension dispatch.
- `mod/objects/src/op_*.go` - query handlers for object storage, reads, scans, search, repo management, probes, type lookup, push, echo, and spec.
- `mod/objects/client/` - typed remote clients used by push, search, create, read, scan, and repository operations.
- `mod/objects/mem/` - in-memory repository implementation for default memory repositories and `objects.new_mem`.
- `mod/objects/fs/`, `mod/objects/views/` - filesystem adapter and presentation helpers.

## Surface

| What | Why it matters |
|---|---|
| `objects.new`, `objects.load`, `objects.store`, `objects.create`, `objects.read`, `objects.delete`, `objects.contains` | core object storage and byte streaming operations |
| `objects.scan`, `objects.search`, `objects.describe`, `objects.probe`, `objects.get_type`, `objects.types`, `objects.spec` | discovery, metadata, type, and inspection operations |
| `objects.push`, `objects.echo` | object delivery and connectivity helpers |
| `objects.repositories`, `objects.remove_repository`, `objects.new_mem` | repository management |
| `Receiver`, `Describer`, `Searcher`, `SearchPreprocessor`, `Finder`, `Holder` | extension points auto-discovered from loaded modules |
| `main`, `device`, `memory`, `local`, `removable`, `virtual`, `network`, `system` | default repository groups and built-in repositories |
| `objects__objects` | object ID to type index used by type lookup and scans by type |

## Invariants

- Every writer must end in exactly one `Commit` or `Discard`; `objects.create` defers `Discard` as a leak guard.
- `objects.delete`, `objects.contains`, and `objects.scan` require an explicit repository.
- `objects.repositories` excludes network routing.
- `Load` returns `*astral.Blob` only for invalid astral magic bytes; other decode failures remain errors.
- Generic object loading rejects sizes above `MaxObjectSize` before reading the object into memory.
- `Drop.Accept(true)` saves at most once even if multiple receivers accept with save.
- Repository scans with `follow=true` must emit exactly one nil between snapshot and live updates.
- `AddRepository` rejects duplicate names; `RemoveRepository` removes the repo from all groups and calls `AfterRemoved` when implemented.
- `ReadDefault` is `main`; `WriteDefault` is `local`.
