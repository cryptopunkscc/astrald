# lib/

Reusable packages sit above raw IPC and below module-specific logic.

## Packages

| Package | Role | Use when |
|---|---|---|
| `lib/astrald` | High-level client. `Client.Query`/`QueryChannel` build an `astral.Query` and call the apphost router; `WithTarget(id)` returns a client clone targeting a remote identity. `NewContext()` returns a guest-identity context with `ZoneAll`. | Call module ops from in-process code or external apps. |
| `lib/apphost` | App IPC primitives: `Connect`, `AuthToken`, `Bind`, `RegisterHandler`, GuestID/HostID, plus `Router` that adapts a `*Host` to `astral.Router`. | Drive IPC directly or build a custom client. |
| `lib/ipc` | IPC transport layer (Unix socket, TCP, in-memory `Conn`). | Plug a non-default IPC transport. |
| `lib/routing` | Merged ops + router primitives. `Op` wraps a `func(*astral.Context, *IncomingQuery[, args]) error`. `OpRouter` maps op names to `Op`s and exposes `AddStruct(s)` / `AddStructPrefix(s, "Op")`. `ScopeRouter` strips the leading `<scope>.` and forwards to a per-scope sub-router (or `root`). `App` = `ScopeRouter` over an `OpRouter` plus a built-in `.spec` op. `PriorityRouter` tries entries in priority order, returns on `nil` error or `ErrRejected`, otherwise falls through to `RouteNotFound`. `IncomingQuery.AcceptRaw()` returns `io.ReadWriteCloser`; `Accept(cfg...)` wraps it as `*channel.Channel`. | Define ops, compose scoped routers, or build a standalone app. |
| `lib/apps` | External app registration over apphost. `Serve(ctx, router, opts...)` creates a handler, registers via the default `AppRegistrar`, gates queries until ready, then blocks routing inbound queries through `router` until ctx is cancelled. `ServeWith(ctx, router, reg, opts...)` uses an explicit `Registrar`. `WithObjectFinder/Searcher/Describer(impl)` mounts the matching op on a `routing.ScopedOpRouter` and adds the replayable `objects.register_*` hook. | Run an external app that the node routes inbound queries to. |
| `lib/query` | Query construction, parsing, and routing helpers. `query.New`, `query.Parse`, `query.RouteInFlight`, `query.Accept(q, src, handler)`, `query.Reject()` and `query.RouteNotFound()` (no arguments). Field tags parsed by `ParseTag` (`required`, `optional`, `skip`, `key:<name>`). | Construct queries, write a custom `Router`, or parse arg tags. |

## Selection Rule

- Caller code uses `lib/astrald` (or a module's `client/` wrapper).
- Handler code uses `lib/routing` (`OpRouter`, `IncomingQuery`).
- Custom routers use `lib/query` helpers and `astral.Router`.
- External apps use `lib/apps` (`Serve` / `ServeWith`).
