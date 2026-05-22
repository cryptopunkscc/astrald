# mod/apphost

Bridges local applications into the node over IPC and HTTP, letting them issue queries as an identity and serve queries addressed to that identity. Owns app access tokens, local app installation records, hosted query handlers, and the contract glue that lets app traffic route through node relays.

## Dependencies

| Module | Why |
|---|---|
| `auth` | authorizes caller override with `SudoAction`, authorizes HTTP object reads with `ReadObjectAction`, signs/indexes app contracts, and searches signed contracts in the query preprocessor |
| `dir` | resolves identities from configured static tokens and HTTP `@alias/path` targets; formats the host alias in `HostInfoMsg` |
| `objects` | stores signed app contracts, serves `Objects.ReadDefault()` through HTTP `/.objects/<id>`, and registers apphost as a purge holder provider |
| `user` (opt) | `PushToLocalSwarm` publishes signed app contracts after sign/install; current sign/install paths call it without a nil guard |
| `core/assets` | `Database()` backs `apphost__access_tokens` and `apphost__local_apps`; `LoadYAML` loads apphost config |

## Flows

- IPC listen: `Run` -> `listen(config.Listen)` opens IPC listeners -> worker goroutines accept conns -> `NewGuest(...).Serve`.
- Guest handshake: `Guest.Serve` sends `HostInfoMsg` -> guest may send `AuthTokenMsg` -> `AuthenticateToken` loads token row -> `AuthSuccessMsg`.
- Guest query: `RouteQueryMsg` -> deny unauthenticated guest when `!AllowAnonymous` -> require `SudoAction` when `Caller` differs from token identity -> apply requested zone and filters -> unauthenticated guest loses `ZoneNetwork` -> `query.RouteInFlight` -> `QueryAcceptedMsg` and `streams.Join`.
- Cancel guest query: `apphost.cancel` finds nonce in `enRoute` -> cancels with optional cause -> sends `Ack`; missing nonce returns query-not-found error.
- Register handler through op: local `apphost.register_handler` -> add `IPCHandler{Identity:q.Caller(), IpcToken, Endpoint}` to `ipcHandlers` -> send `Ack`.
- Bind handler lifetime: local `apphost.bind` sends `Ack` -> consumes `BindMsg` tokens -> on close removes matching IPC handlers.
- Route inbound query to app: `Module.RouteQuery` scans `ipcHandlers` for `q.Target` -> `ipc.DialContext(handler.Endpoint)` -> send `HandleQueryMsg` -> app `Ack` accepts, `QueryRejectedMsg` rejects, unavailable endpoint is removed.
- Query preprocessing: find local app hosting contract for `Caller` and attach it -> find relay contracts issued by `Target` -> add non-local contract subjects as relays.
- Create/list tokens: `apphost.create_token` writes random 32-char token with default 1-year expiry when duration is empty; `apphost.list_tokens` streams matching tokens and ends with `EOS`.
- Static tokens: `Prepare` resolves each `config.Tokens` identity -> deletes existing token row -> inserts a new token with a 100-year expiry.
- Install app: local `apphost.install_app` -> build `RelayForAction` contract for app and node -> `Auth.SignContract` -> `Auth.IndexContract` -> `Objects.Store` -> `CreateLocalApp` upsert -> async `User.PushToLocalSwarm`.
- Object holds: local `apphost.hold_object` inserts an app-owned hold with an optional `duration` arg, `apphost.unhold_object` deletes only the caller-owned hold, and `apphost.list_held_objects` streams the caller's active holds. `objects.purge` skips any object with at least one active apphost hold. Active means `hold_until IS NULL OR hold_until > now`; omitted duration writes `NULL` (infinite hold), otherwise `hold_until = now + duration`.
- HTTP bridge: bearer token -> `AuthenticateToken` -> set guest/host headers -> `/.objects/<id>` goes through `ReadObjectAction`; other paths route a query to header target, `@alias/path` target, or local node.

## Source

- `mod/apphost/module.go`, `contract.go`, `access_token.go`, `app.go` - public module interface, method names, app contract builder, token/app objects.
- `mod/apphost/*_msg.go` - IPC and apphost wire messages exchanged with local apps.
- `mod/apphost/src/loader.go`, `module.go`, `deps.go`, `prepare.go`, `config.go` - config, database setup, dependency injection, static-token sync, listeners/workers startup.
- `mod/apphost/src/listen.go`, `worker.go`, `guest.go` - IPC listener fan-in, worker loop, guest authentication and outbound query protocol.
- `mod/apphost/src/query_router.go`, `ipc_handler.go`, `query_preprocessor.go` - hosted app dispatch and relay-contract preprocessing.
- `mod/apphost/src/http_server.go`, `http_object_handler.go`, `http_query_handler.go` - bearer-auth HTTP bridge for objects and queries.
- `mod/apphost/src/op_*.go` - query operation handlers.
- `mod/apphost/src/access_tokens.go`, `db.go`, `db_access_token.go`, `db_local_app.go`, `db_object_hold.go` - token, local-app, and object-hold persistence plus lookup helpers.
- `mod/apphost/src/object_holder.go` - objects holder hook backed by active app-owned hold rows.
- `mod/apphost/client/` - typed client wrappers for apphost ops.

## Surface

| What                                                                           | Why it matters                                                                               |
|--------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------|
| `apphost.create_token`, `apphost.list_tokens`                                  | issue and stream app access tokens                                                           |
| `apphost.hold_object`, `apphost.unhold_object`, `apphost.list_held_objects`     | manage persistent app-owned object holds                                                     |
| `objects.Holder`                                                               | prevents actively held app-owned objects from being purged                                   |
| `apphost.register_handler`, `apphost.bind`, `apphost.cancel`                   | manage hosted app handlers and cancel in-flight guest queries                                |
| `apphost.new_app_contract`, `apphost.sign_app_contract`, `apphost.install_app` | create, sign, store, and index relay contracts for local apps                                |
| `Module.RouteQuery`                                                            | high-priority router that forwards target-addressed queries to registered local app handlers |
| `Module.PreprocessQuery`                                                       | attaches local app relay contracts and adds relay hops for apps hosted elsewhere             |
| `config.Listen`, `config.BindHTTP`                                             | configure IPC apphost listeners and the HTTP bridge                                          |
| `apphost__access_tokens`, `apphost__local_apps`, `apphost__object_holds`        | persistent access tokens, installed-app records, and app-owned object holds                  |

## Invariants

- `apphost.install_app`, `apphost.register_handler`, and `apphost.bind` reject network-origin queries.
- `apphost.hold_object`, `apphost.unhold_object`, and `apphost.list_held_objects` reject network-origin queries and require a non-zero caller identity.
- Apps can list and unhold only their own hold rows; many apps may hold the same object.
- Anonymous IPC guests can route only when `allow_anonymous` is true and always lose `ZoneNetwork`.
- A guest can set `Caller` to another identity only when `auth.Authorize(SudoAction{Actor:guestID, AsID:Caller})` grants it.
- `apphost.list_tokens` always ends the stream with `EOS`.
- `CreateLocalApp` uses `OnConflict{DoNothing}`; reinstall does not update an existing row.
- `HoldObject` uses `OnConflict{DoNothing}`; duplicate holds are idempotent, holding does not require the object to exist locally, and `Duration` is `*astral.Duration` (nil = no expiry).
- `bind_http` empty disables the HTTP bridge.
- Handler endpoints are removed when dial fails during inbound routing.
