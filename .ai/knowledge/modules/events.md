# mod/events

Wraps arbitrary local data objects into `events.Event` objects and submits them to the object receiver for local fan-out. Owns only the event object shape and the `Emit` helper; subscription and delivery behavior live in `objects`.

## Dependencies

| Module | Why |
| --- | --- |
| `objects` | `Emit` calls `Objects.Receive(event, localID)` asynchronously so local receivers can observe the event |

## Flows

- Emit event: `Emit(data)` builds an `Event` with a fresh nonce, local node identity, current time, and the original data object -> starts `Objects.Receive` in a goroutine -> returns the event pointer immediately.
- Lifecycle: `Run` does no event work and waits for context cancellation; all behavior is driven by `Emit`.

## Source

- `mod/events/module.go`, `event.go` - public module interface and event wire object.
- `mod/events/src/loader.go`, `module.go`, `deps.go`, `config.go` - registration, dependency injection, `Emit`, and lifecycle.

## Invariants

- Fire-and-forget; receive error dropped.
- `SourceID` is local identity at emit time.
- `Data` referenced, not cloned.
- No ordering across concurrent emits.
