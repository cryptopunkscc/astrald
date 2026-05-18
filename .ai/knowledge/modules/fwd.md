# mod/fwd

Bridges byte streams between configured endpoints by accepting traffic on a local server and routing it to an astral, TCP, or Tor target. Owns static forwarding setup, per-forward server runners, and the protocol adapters that translate accepted streams into routed queries or outbound sockets.

## Dependencies

| Module | Why |
|---|---|
| `astral.Node` | supplies local identity and routes astral target queries through `node.RouteQuery` |
| `dir` | resolves identity names in `astral://caller@target:query` targets and formats display names for labels |
| `tcp` (opt) | injected in `Deps`; current forwarding code uses Go `net` directly instead |
| `tor` (opt) | required for `tor://` targets; parses target endpoints and dials Tor sockets |
| `core/assets` | `LoadYAML` reads the `fwd` config and its `Forwards` map |

## Flows

- Startup forwards: `Run` includes `ZoneNetwork` in the module context -> iterates `config.Forwards` -> `CreateForward(server, target)` logs and skips entries that fail to parse or start.
- Forward creation: `CreateForward` parses the target URI into an `astral.Router` -> parses the server URI -> wraps the server in `ServerRunner` -> `runServer` tracks it and starts its goroutine.
- TCP inbound: `NewTCPServer` opens `net.Listen("tcp", bind)` -> `Run` accepts clients until context closes the listener -> each client launches a blank query through `target.RouteQuery` -> copies bytes from client to returned writer.
- Astral target: `parseTarget("astral://...")` resolves optional caller and target identities -> builds a template query -> `AstralTarget.RouteQuery` clones it, refreshes the nonce, and routes it through the node.
- TCP target: `TCPTarget.RouteQuery` dials the configured TCP address -> starts remote-to-caller copy in a goroutine -> returns the TCP connection for caller-to-remote copy; dial failure rejects the query.
- Tor target: `TorTarget.RouteQuery` parses and dials through `mod/tor` -> starts remote-to-caller copy -> returns the Tor connection; parse or dial failure rejects the query.
- Astral server path: `NewAstralServer` can parse identity-qualified service names, but `AstralServer.Run` currently returns `obsolete`, so configured astral inbound forwards fail after startup.
- Shutdown: `Run` waits for context cancellation -> `waitForServers` waits on active `ServerRunner.Done` channels; each runner removes itself from `Module.servers` after its server returns.

## Source

- `mod/fwd/src/loader.go`, `mod/fwd/src/deps.go`, `mod/fwd/src/config.go` - module registration, dependency injection, and configured forward map.
- `mod/fwd/src/module.go` - startup loop, URI parsing, server creation, and active server tracking.
- `mod/fwd/src/server.go` - `Server` interface and `ServerRunner` lifecycle wrapper.
- `mod/fwd/src/tcp_server.go`, `mod/fwd/src/astral_server.go` - inbound TCP and obsolete astral server implementations.
- `mod/fwd/src/tcp_target.go`, `mod/fwd/src/tor_target.go`, `mod/fwd/src/astral_target.go` - outbound target adapters.
- `mod/fwd/src/README.md` - local source note for forwarder package context.

## Surface

| What | Why it matters |
|---|---|
| `config.Forwards` | static mapping from server URI to target URI |
| `TCPServer` | only functional inbound listener in current source |
| `AstralTarget`, `TCPTarget`, `TorTarget` | outbound adapters that satisfy `astral.Router` |
| `Module.Servers()` | snapshot of active server runners |

## Invariants

- Server and target URIs must contain `://`; unsupported protocols fail before a runner starts.
- `mod.ctx` includes `ZoneNetwork` for all configured forwards.
- `tor://` targets require `mod.Tor`; missing Tor support makes target parsing fail.
- TCP and Tor dial failures call `query.Reject()` and return that rejection error.
- `AstralServer.Run` is intentionally non-functional and returns `obsolete`.
- `ServerRunner.Stop` only cancels its context; protocol cleanup happens in the concrete server.
- TCP inbound creates one goroutine per accepted client and does not enforce a connection cap.
