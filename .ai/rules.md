# Rules

## Work Discipline

- Read relevant code before editing.
- Keep changes scoped to the task.
- Preserve user changes; never revert unrelated work.
- Call out conflicts between code, docs, and `.ai` context.
- Verify behavior with focused tests/checks when code changes.

## Context Discipline

- Keep default context small.
- Use indexes before loading scoped files.
- Ignore `.ai/artifacts/` unless explicitly referenced.
- Treat AI working notes as provisional until promoted.
- Correct stale `.ai` context when found.

## Engineering Judgment

- Design around data, invariants, and state transitions.
- Prefer explicit state over hidden control flow.
- Reduce special cases instead of layering branches.
- Use the standard library first.
- Add abstraction on third use, only for same algorithm/data flow.

## Code Shape

- Functions: one responsibility, max 50 lines, max 4 params.
- 3+ related returns -> named struct.
- Packages: one concept. No `util`, `common`, `helpers`.
- Interfaces live at consumers. Prefer 1 method; 3+ is suspect.
- Flat structs. `nil` as sentinel.
- 2 options -> 2 params. Rare 5+ options -> options struct.

## Code Style

- Naming: precise verbs, e.g. `delete`, `find`, `create`.
- Logging: always `%v`. Levels: `Log` 0, `Logv(1)` verbose, `Logv(2)` debug.
- Annotate code as you write it per `comments.md` (always-on): `// todo` `// fixme` `// note` `// why`. Intent, not mechanics.

## Project APIs

- Use `astral.Objectify` for `WriteTo`/`ReadFrom`.
- Objectify fields must be astral primitives; plain Go fields are not handled.
- Use `astral.Adapt(v)` to wrap a native Go value into an astral `Object`; do not hand-roll switch ladders. Pass-through for `Object`, `nil`→`&Nil{}`, `error`→`NewError`. Default widths: `int`/`uint`→`Int64`/`Uint64`, `string`→`String32`. When the spec dictates a narrower width (e.g. `uint16`, `string16`), Adapt is the wrong tool — dispatch on the spec first.
- Use `objects.Save`/`objects.Load`, not raw `WriteTo`.
- Inject dependencies with `core.Inject(node, &mod.Deps)` in `LoadDependencies`.
- Prefer `sig.Map`/`sig.Set`/`sig.Queue` over mutex + map/slice.
- Use `sig.RecvErr`/`sig.Recv`/`sig.Send` for context-aware channel ops.

## Domain Invariants

- Every repository writer must `Commit()` or `Discard()`.
- Never access other modules during `Load`.
- Zones narrow only; never expand at a hop.
- Check `ctx.Zone().Is(astral.ZoneNetwork)` before network work.
- Default context is Device|Virtual; original caller must add Network.
- `query.Reject()` is terminal.
- `query.RouteNotFound(r, ...)` is non-terminal.
- Never return `nil, nil` from `RouteQuery`.
- Streaming ops end with `ch.Send(&astral.EOS{})`.
- Send stream errors with `ch.Send(astral.Err(err))`.

## Concurrency

- Mutex field name: `mu`; never embed.
- Put `defer Unlock()` on the same line as `Lock()`.
- Use `sync.RWMutex` when reads dominate.
- Atomics: `Bool` for flags, `Int32` for states, `Uint64` for counters.
- Idempotent close uses `CompareAndSwap(false, true)`.
- Do not use `sync.Once`; use `atomic.Bool.CompareAndSwap`.
- `sync.Cond` only for computed blocking conditions; `.Wait()` in `for`.
- Simple done/ready signal -> channel.
- `wg.Add(1)` before `go`; `defer wg.Done()` first in goroutine.
- WaitGroups are local variables, never struct fields.
- Signal done by `close()`, not send. Expose `<-chan struct{}`.
- `sig.Sig` is canonical read-only signal; `sig.New()` is buffered(1).
- Error channels are buffered with capacity >= senders.
- `<-ctx.Done()` must be in `select` with another case.

