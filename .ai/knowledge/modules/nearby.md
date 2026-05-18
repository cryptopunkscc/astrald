# mod/nearby

Discovers peers on the local network by exchanging status and scan messages over the ether broadcast path. Owns local discovery mode, status composition, the lazy-expiring cache of observed peers by source IP, and resolver hooks that make nearby aliases and endpoints visible to `dir` and `nodes`.

## Dependencies

| Module | Why |
|---|---|
| `ether` | sends `ScanMessage` and `StatusMessage` with `Push` and `PushToIP`; receives `EventBroadcastReceived` through objects |
| `objects` | registers `ReceiveObject`; returns `ErrPushRejected` for broadcast objects that should not be stored |
| `user` | `Run` waits for `Ready`; `Mode` falls back to visible before identity exists; stealth resolution uses the local user identity |
| `dir` | `AddResolver` registers nearby alias lookup and `GetAlias` contributes the local alias to composed status |
| `nodes` | `AddResolver` registers endpoint lookup from cached status attachments |
| `tree` | persists and follows the discovery mode at `/mod/nearby/mode` |
| `auth` | status identity resolution can derive an identity from signed contract attachments |
| `core` | dependency injection and loaded-module scan for `nearby.Composer` implementations |
| `exonet`, `shell`, `tcp` | injected dependencies kept available for composers and local integrations |

## Flows

- Startup: `Run` waits for `User.Ready` -> applies YAML `mode` to the tree value when present -> starts periodic updater -> sends an initial scan -> blocks on context.
- Periodic broadcast: mode updates or `statusExpiration - 5s` timer -> `Broadcast` skips silent mode -> `Status(nil)` asks each composer for attachments -> stealth without attachments is suppressed -> `Ether.Push`.
- Scan request: local ether broadcast event carrying `ScanMessage` -> compose status for `astral.Anyone` -> `Ether.PushToIP` back to the source IP when broadcasting is allowed.
- Inbound status: local self-echoed ether event carrying `StatusMessage` -> cache by source IP with timestamp -> async `ResolveStatus` fills identity when possible -> replace cache entry.
- Network address change: local IP address-change event -> push current status -> send a fresh scan.
- Identity resolution: public profile attachment wins -> signed contract subject can identify the node -> stealth hint recomputes commitment from known user identity and nonce, then unmasks the node identity.
- Endpoint lookup: `ResolveEndpoints` clones the live cache -> filters entries by resolved identity -> streams `*nodes.EndpointWithTTL` attachments.
- Alias lookup: resolver accepts names with `.` prefix -> strips the prefix -> matches `*dir.Alias` attachments in cached statuses.
- List operation: `nearby.list` snapshots `Cache()` -> streams cached statuses -> sends `EOS`.
- Broadcast operation: `nearby.broadcast` triggers `Broadcast` immediately and returns `Ack` or an error object.

## Source

- `mod/nearby/module.go`, `mod/nearby/mode.go`, `mod/nearby/status_message.go`, `mod/nearby/scan_message.go`, `mod/nearby/status.go`, `mod/nearby/public_profile.go`, `mod/nearby/stealth_hint.go`, `mod/nearby/flag.go`, `mod/nearby/errors.go` - public interfaces, modes, wire objects, and identity-hiding helpers.
- `mod/nearby/src/loader.go`, `mod/nearby/src/deps.go`, `mod/nearby/src/config.go`, `mod/nearby/src/module.go` - configuration, dependency registration, mode persistence, cache, scan, and broadcast loop.
- `mod/nearby/src/composition.go` - composer-facing attachment builder and max attachment size enforcement.
- `mod/nearby/src/object_receiver.go` - ether and IP-change event dispatch.
- `mod/nearby/src/resolve_status.go` - identity resolution from profile, contract, and stealth hint attachments.
- `mod/nearby/src/endpoint_resolver.go`, `mod/nearby/src/identity_resolver.go` - integrations with `nodes` endpoint resolution and `dir` identity resolution.
- `mod/nearby/src/op_broadcast.go`, `mod/nearby/src/op_list.go` - query handlers.
- `mod/nearby/views/status_view.go` - status display helper.

## Surface

| What | Why it matters |
|---|---|
| `nearby.broadcast`, `nearby.list` | manual broadcast trigger and cached peer inspection |
| `nearby.Composer` | extension point for modules to attach aliases, endpoints, contracts, and other status facts |
| `dir.Resolver` | resolves dot-prefixed nearby aliases from cached status attachments |
| `nodes.EndpointResolver` | serves endpoint attachments discovered on the local network |
| `objects.Receiver` | consumes local ether broadcast events and local IP address changes |
| `/mod/nearby/mode` | persisted visible, stealth, or silent discovery mode |

## Invariants

- `ReceiveObject` ignores drops not sent by the local node; nearby discovery is driven by ether self-echo events.
- `Cache()` performs lazy eviction for entries older than `statusExpiration`.
- `Mode()` returns visible when the user identity is still nil.
- Stealth mode with zero attachments is suppressed and behaves like silent mode for broadcast.
- Attachments larger than `MaxAttachmentSize` are rejected with `ErrObjectTooLarge`.
- The periodic updater broadcasts five seconds before cache expiry to avoid normal entries timing out.
- YAML `mode` is applied only when configured; otherwise the persisted tree value remains authoritative.
