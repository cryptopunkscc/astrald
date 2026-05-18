# mod/tree

Maintains a path-addressed namespace of astral objects backed by a database node table and live value subscriptions. Owns path traversal, mount points for local or remote nodes, tree query ops, and reflection-based binding of Go structs and typed values to tree paths.

## Dependencies

| Module | Why |
|---|---|
| `dir` | `OpMountRemote` resolves the target identity for remote tree mounts |
| `core/assets` | `LoadYAML` reads config and `Database()` backs the `tree__nodes` table |
| `gorm` | `DB` migrates and queries persisted tree nodes and object payloads |
| `lib/routing` | `OpRouter.AddStructPrefix` exposes tree query handlers |
| `lib/query` | tree client nodes issue `tree.get`, `tree.set`, `tree.delete`, and `tree.list` queries |
| `astral/channel` | tree ops stream values, names, acknowledgements, errors, and `EOS` with negotiated formats |
| `sig` | mount registry uses `sig.Map`; value subscriptions use `sig.Queue` and `sig.Subscribe` |

## Flows

- Module setup: `Load` reads config -> registers `Op*` handlers -> opens the database wrapper -> migrates `tree__nodes` -> mounts the DB root at `/`.
- Path traversal: `tree.Query` walks path segments through `Sub`; `create=true` calls `Create` for missing segments and treats `ErrAlreadyExists` as a concurrent-create retry signal.
- Mount splicing: `Root` returns a `NodeWrapper`; `NodeWrapper.Sub` checks each child absolute path against `mod.mounts` and substitutes mounted nodes before returning wrapped children.
- Get value: `tree.get` delegates to `NodeOps.Get` -> traverses without creation -> `Node.Get` loads the persisted payload or `astral.Nil` -> optionally subscribes to further value updates -> stream ends with `EOS`.
- Set value: `tree.set` traverses with creation -> receives objects from the channel -> nil object becomes `astral.Nil` -> `db.setNodeValue` upserts payload -> `pushNodeValue` notifies subscribers -> sends `Ack`.
- Delete node: `tree.delete` traverses without creation -> rejects nodes with children using `ErrNodeHasSubnodes` -> closes subscriber queue if present -> deletes the database row -> sends `Ack`.
- List children: `tree.list` defaults empty path to `/` -> traverses without creation -> streams child names as `String8` -> sends `EOS`.
- Remote mount: `tree.mount_remote` resolves target identity -> creates `tree/client.Node` for that target -> optionally traverses the remote root path -> inserts it into `mod.mounts`; `tree.unmount` removes the absolute mount path.
- Remote proxy node: client `Node.Get`, `Set`, `Delete`, `Sub`, and `Create` translate tree node operations into query-channel calls against a target identity.
- Struct and value binding: `Bind` walks exported struct fields -> uses `tree` tag path or snake_case field name -> either calls field `Bind` or recurses; `Value.Bind` follows the node with `Get(follow=true)` and caches updates.

## Source

- `mod/tree/module.go`, `node.go`, `bind.go`, `value.go`, `nil_node.go`, `errors.go`, `err_no_value.go` - public interfaces, traversal helpers, bind support, typed values, nil node, and errors.
- `mod/tree/src/loader.go`, `module.go`, `deps.go`, `config.go` - construction, root mount, database setup, dependency injection, and lifecycle.
- `mod/tree/src/node.go`, `node_wrapper.go`, `db.go`, `db_node.go` - DB-backed nodes, mount substitution, persistence helpers, and schema.
- `mod/tree/src/op_get.go`, `op_set.go`, `op_delete.go`, `op_list.go`, `op_mount_remote.go`, `op_unmount.go` - query handlers.
- `mod/tree/client/client.go`, `node.go`, `server.go` - remote tree client and shared node operation handlers.

## Surface

| What | Why it matters |
|---|---|
| `tree.get`, `tree.set`, `tree.delete`, `tree.list` | path CRUD query surface, including follow-mode reads |
| `tree.mount_remote`, `tree.unmount` | exposes remote subtree mounting and mount removal |
| `tree.Module` and `tree.Node` | core interfaces used by modules that bind config or expose custom tree nodes |
| `tree.Bind`, `tree.BindPath`, `tree.Value` | reflection and typed-value layer used for tree-backed configuration |
| `tree__nodes` | persisted tree namespace and object payload table |

## Invariants

- The root DB node cannot hold a value.
- Paths mounted through `Mount` and `Unmount` must be absolute; trailing slashes are trimmed.
- One mount can exist per normalized path; duplicate mount and missing unmount return errors.
- Nodes with children cannot be deleted.
- Missing node values are represented as `astral.Nil`; nil values written through `Set` are stored as `astral.Nil`.
- `pushNodeValue` is a no-op until a subscriber exists; deleting a node closes and removes its subscriber queue.
- Unknown stored object blueprints decode as `astral.UnparsedObject` through the DB helper.
- Tree streaming ops that enumerate values or child names terminate with `EOS`.
