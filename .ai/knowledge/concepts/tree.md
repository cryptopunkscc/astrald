# Tree

`mod/tree` is a hierarchical key-value store.

* Every node can hold an `astral.Object`.
* Every node can have named children.
* Paths use the `/segment/segment` convention.
* `Mount` can replace any subtree with a custom `Node` implementation.
* `MountRemote` can mount a remote node from another Astral identity.

The tree is network-addressable. An operator on another machine can write
values into a peer's tree over an encrypted Astral connection. Treat the tree as
the node's shared runtime state layer, not only as local config.

## Live Config Binding

Modules bind settings to tree paths with `tree.Value[T]`.

`tree.Value[T]` is a typed, persistent, observable cell:

* It holds the current value.
* It survives restarts through the DB.
* It notifies watchers on change.

### Read APIs

* `Get()` performs a one-shot read of the current value.
* `Follow(ctx)` returns a channel subscription. It delivers the current value
  immediately, then every future change.

Use `Follow(ctx)` when module behavior must react continuously. For example,
`mod/tcp` toggles its listener goroutine on or off as `settings.Listen`
changes.

### Invariants

* A module cannot refuse a new value from `Follow(ctx)`.
* A module reacts to whatever value arrives.

### Paths

* Use `/mod/<name>/config` for persistent config.
* Use `/mod/<name>/settings` for runtime-togglable state.
* `Bind()` wires a struct's `tree.Value` fields automatically, using field
  names as path segments.
