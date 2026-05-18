# mod/fs

Exposes local directories as `objects.Repository` instances so the node can store, read, describe, and search content-addressed objects from disk. Owns writable filesystem repositories, read-only watched repositories, the local-file index, and fsnotify-driven reindexing for paths under watched roots.

## Dependencies

| Module | Why |
|---|---|
| `objects` | registers repositories with `AddRepository`, adds them to `RepoLocal`, implements `Repository`, `AfterRemovedCallback`, `Describer`, and `Searcher` contracts |
| `core/assets` | `LoadYAML` reads fs config, `Database()` backs `fs__local_files`, and file resources provide the default `<data root>/data` repo |
| `auth` (opt) | injected in `Deps`; current source does not call it |
| `dir` (opt) | injected in `Deps`; current source does not call it |
| `shell` (opt) | injected in `Deps`; current source does not call it |
| `fsnotify` | drives `Watcher` events for writes, renames, creates, chmods, and removes |
| `golang.org/x/time/rate` | bounds stat, hash, and enqueue work in `Indexer` |
| `lib/paths` | walks watched roots, collapses overlapping roots, checks path coverage, and tracks path prefixes |

## Flows

- Default repository: `LoadDependencies` -> `addDefaultRepo` -> require file-backed resources -> create `<DataRoot>/data` -> `NewRepository` -> `Objects.AddRepository("data")` -> `Objects.AddGroup(RepoLocal, "data")`.
- Configured repositories: iterate `config.Repos` -> `Writable` selects `NewRepository`, otherwise `NewWatchRepository` -> register with `objects` -> add to `RepoLocal`.
- Writable object commit: `Repository.Create` checks requested allocation against `DiskUsage` -> `NewWriter` writes to `.tmp.<hex>` -> `Commit` resolves object ID -> atomic finalized compare-and-swap -> rename temp file to the object ID -> publish to `Repository.addQueue`.
- Writable scan with follow: subscribe to `addQueue` first -> list root directory -> emit parseable object ID filenames -> send a single `nil` snapshot boundary -> forward later queue events.
- Watch repository construction: reject non-absolute or non-directory root -> create `Watcher` -> map write and rename to `onChange`, remove to `onRemove`, new directories to recursive `Add` -> add root to watcher and indexer.
- Watch repository removal: `AfterRemoved` cancels active scan, closes watcher, and removes the root from `Indexer`; uncovered indexed paths are soft-deleted.
- Watch read and scan: `Read` looks up active rows by root and object ID, opens the first readable path, and returns a bounded `Reader`; `Scan(follow=true)` emits indexed IDs, a `nil` boundary, then index events under the root.
- Indexer startup: `Run` starts four workers -> `init` walks widest roots -> validates unchanged rows -> invalidates missing or stale rows -> enqueues invalidated paths through the rate-limited queue.
- Indexer worker: pop path -> clear pending flag -> soft-delete uncovered paths -> stat covered paths -> validate unchanged rows -> hash changed files with `astral.Resolve` -> upsert index row -> publish `IndexEvent` when subscribers exist.
- Filesystem changes: debounced write and rename events call `requeuePath`; remove events hard-delete the path row; directory creation adds a recursive watcher.
- Repository ops: `fs.new_repo` creates a writable repository and registers it; `fs.new_watch` creates a watched repository, starts an index scan, registers it, and cancels the scan if registration fails.
- Describe and search: `DescribeObject` and `SearchObject` require `ZoneDevice`; search also requires the `path` tag and matches indexed paths case-insensitively.

## Source

- `mod/fs/module.go`, `mod/fs/errors.go`, `mod/fs/events.go`, `mod/fs/file_location.go` - public module name, errors, event types, and file-location descriptor object.
- `mod/fs/src/loader.go`, `mod/fs/src/module.go`, `mod/fs/src/deps.go`, `mod/fs/src/config.go` - lifecycle, dependency injection, YAML config, database setup, default and configured repository registration.
- `mod/fs/src/op_new_repo.go`, `mod/fs/src/op_new_watch.go` - query handlers for runtime repository registration.
- `mod/fs/src/repository.go`, `mod/fs/src/writer.go`, `mod/fs/src/reader.go` - writable repository scan, read, create, commit, discard, delete, and free-space behavior.
- `mod/fs/src/watch_repository.go`, `mod/fs/src/watcher.go` - read-only watched repositories and fsnotify event adaptation.
- `mod/fs/src/indexer.go` - root tracking, worker queue, scan, validation, hashing, soft-delete, and subscription behavior.
- `mod/fs/src/db.go`, `mod/fs/src/db_local_file.go`, `mod/fs/src/batch_collector.go` - GORM schema and batched index queries.
- `mod/fs/src/disk_usage.go` - filesystem free-space probe used before writable creates.
- `mod/fs/src/object_describer.go`, `mod/fs/src/object_searcher.go` - object descriptor and path search extension points.
- `mod/fs/views/file_location_view.go` - terminal rendering for `FileLocation`.

## Surface

| What | Why it matters |
|---|---|
| `fs.new_repo`, `fs.new_watch` | runtime creation of writable and watched local repositories |
| `Repository` | writable content-addressed object store backed by files named as object IDs |
| `WatchRepository` | read-only repository over an existing absolute directory tree |
| `Indexer` | durable path-to-object index and fsnotify recheck queue for watched roots |
| `fs__local_files` | persistent index of paths, object IDs, modification times, update state, and deletion state |
| `DescribeObject`, `SearchObject` | object metadata and path lookup hooks exposed to `objects` |

## Invariants

- Writable repository files are named exactly `ObjectID.String()`; non-regular and unparseable filenames are skipped during scan.
- `Repository.Read`, `DescribeObject`, and `SearchObject` require `ZoneDevice`; `WatchRepository.Read` does not check the zone.
- `WatchRepository` is read-only: `Create` and `Delete` return `errors.ErrUnsupported`, and `Free` returns `0`.
- Watch roots are cleaned absolute directories that must exist at construction time.
- Active index rows require `updated_at != 0` and `deleted_at IS NULL`; `updated_at = 0` means the path needs recheck.
- `Writer` finalization is single-use; repeated `Commit` or `Discard` returns `writer closed`.
- Follow scans emit one `nil` object ID as the snapshot/follow boundary.
- `IndexEvent` fan-out is active only while `subscriberCount > 0`.
- Removing a watch root soft-deletes only paths not covered by another remaining root.
- `Repository.Create` rejects with `objects.ErrNoSpaceLeft` when `opts.Alloc` exceeds reported free space.
