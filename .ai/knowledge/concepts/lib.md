# lib/

Reusable packages sit above raw IPC and below module-specific logic.

## Packages

| Package | Role | Use when |
|---|---|---|
| `lib/astrald` | High-level client. Queries any Identity through the local node. Uses `astrald.Default()` via apphost. | Call module ops. |
| `lib/apphost` | Low-level App IPC: handshake, token auth, GuestID, routing, handler registration. | Control IPC directly. |
| `lib/ipc` | IPC transport: Unix sockets, TCP, in-memory. | Change IPC transport. |
| `lib/routing` | Handler routing. `OpRouter`, `Op`, and `IncomingQuery` map `Op*` methods to op strings, parse args, and dispatch to handler goroutines. `IncomingQuery.AcceptRaw()` returns raw `io.ReadWriteCloser`; `Accept(cfg...)` returns `*channel.Channel`. `ScopeRouter` strips the dot-prefix before forwarding. `App` is `ScopeRouter` + `OpRouter` with built-in `.spec` op. `PriorityRouter` tries routers by priority and stops on Accept or Reject. | Expose module ops, compose scoped sub-routers, or build standalone apps. |
| `lib/apps` | External app registration. `Serve()` and `AppRegistrar` register IPC handlers with the node and gate queries until ready. | Register external app handlers. |
| `lib/query` | Query construction and parsing. Includes `RouteNotFound`, `Reject`, and `Accept` helpers. | Route or build queries. |

## Selection Rule

Use `lib/astrald` for callers, `lib/routing` for handlers, `lib/query` for routers, and `lib/apps` for external app registration.
