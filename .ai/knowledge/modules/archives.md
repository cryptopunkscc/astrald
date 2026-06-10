# mod/archives

Indexes ZIP archives stored as objects so their entries become individually addressable, searchable, and readable through the objects pipeline. Owns the persistent archive-and-entry index, the virtual-zone object opener for entries, the device-zone archive descriptor, and the per-entry read-authorization rule that delegates to the parent archive.

## Dependencies

| Module | Why |
|---|---|
| `objects` | implements `objects.Describer` for `ArchiveDescriptor`, exposes entries via `OpenObject`, serves `objects.SearchObject`, calls `Objects.Receive` to emit `EventArchiveIndexed`, and reads parent archive bytes through `Objects.ReadDefault` |
| `auth` | registers `AuthorizeObjectsRead` for `objects.ReadObjectAction` and recursively calls `Auth.Authorize` on the parent archive to decide entry access |
| `core` | module registration and dependency injection |

## Flows

- Index: `Index(ctx, zipID)` locks `mod.mu` -> returns cached `Archive` from `dbArchive`+`dbEntry` if present -> otherwise `scan` opens the zip through `readerAt`+`objects.ReadDefault` -> `astral.Resolve` each file to get an entry `ObjectID` -> `setCache` clears and rewrites rows -> `Objects.Receive(&EventArchiveIndexed{...})`.
- Forget: `Forget(zipID)` calls `clearCache` which deletes `dbEntry` rows by parent then the `dbArchive` row.
- Describe: `DescribeObject` rejects non-device zones -> reads cached archive -> sums entry sizes -> emits one `Descriptor{Data: ArchiveDescriptor{Format, Entries, TotalSize}}`.
- Open entry: `OpenObject` requires virtual zone -> bounds-checks via `objects.IsOffsetLimitValid` -> looks up `dbEntry` rows for the target objectID -> opens parent zip via `openZip` and returns a `contentReader` over the named path; falls back to `objects.ErrNotFound` if every candidate parent fails.
- Search entries: `SearchObject` requires virtual zone and the `path` or `archive` tag -> `LOWER(path) LIKE %query%` over `dbEntry` -> streams `SearchResult{ObjectID: entry.ObjectID}` until done.
- Authorize entry read: `AuthorizeObjectsRead` finds all `dbEntry` rows for the target ID -> for each, recurses with `Auth.Authorize` on the parent archive's `ReadObjectAction`; denies otherwise.
- Random-access read: `contentReader.Seek` forward-skips via `streams.Skip`, backward-seeks by reopening the zip entry and skipping from start; `readerAt.ReadAt` builds an `astral.Context` with `openTimeout` and calls `Objects.ReadDefault().Read`.

## Source

- `mod/archives/module.go`, `mod/archives/archive_descriptor.go`, `mod/archives/events.go` - public `Module` interface, `Archive`/`Entry` types, the `ArchiveDescriptor` object, and the `EventArchiveIndexed` object.
- `mod/archives/src/loader.go`, `mod/archives/src/module.go`, `mod/archives/src/deps.go`, `mod/archives/src/config.go` - registration, YAML config, GORM auto-migration, dependency wiring, and `auth` rule installation.
- `mod/archives/src/db.go` - `dbArchive` and `dbEntry` tables with `archives__` prefix and `OnDelete:CASCADE` between them.
- `mod/archives/src/index.go` - `Index`, `scan`, `Forget`, and cache get/set/clear.
- `mod/archives/src/object_describer.go`, `mod/archives/src/object_searcher.go`, `mod/archives/src/object_opener.go` - zone-gated describer, searcher, and entry opener.
- `mod/archives/src/authorizer.go` - recursive read-authorization rule for entry IDs.
- `mod/archives/src/content_reader.go`, `mod/archives/src/reader_at.go` - seekable zip-entry reader and the `io.ReaderAt` adapter over `objects.ReadDefault`.

## Invariants

- `Index` is serialized by `mod.mu`; concurrent indexing of the same archive cannot race.
- `Describe` is device-zone only; `OpenObject` and `SearchObject` are virtual-zone only.
- `setCache` clears existing rows first, so re-indexing is idempotent and replaces stale entries.
- `AuthorizeObjectsRead` returns false when the entry's parent row is missing or when the entry ID equals the parent ID (sanity guard against self-reference).
- `SearchObject` requires the caller's `RequiredTags` to be a subset of `{path, archive}`.
