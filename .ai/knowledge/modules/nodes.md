# mod/nodes

Establishes and maintains authenticated encrypted links to peer nodes, then multiplexes query sessions across those links for network-zone routing. Owns endpoint persistence, link strategy activation, Noise and mux negotiation, flow-controlled session I/O, relayed query authorization, and session migration between links.

## Dependencies

| Module | Why |
|---|---|
| `exonet` | dials and parses transport endpoints for basic, tor, and NAT-backed links |
| `crypto` | provides the local secp256k1 private key used by the Noise XK handshake |
| `dir` | resolves identities in ops and registers the `linked` identity filter with `SetFilter` |
| `auth` | registers `RelayForAction` authorizer and authorizes inbound `RelayQuery` frames |
| `objects` | pushes caller proofs for relayed routing and registers describer, finder, and receiver integrations |
| `scheduler` | schedules create-link, ensure-link, and endpoint-cleanup tasks |
| `events` | emits link lifecycle, link pressure, and observed endpoint events |
| `user` | `LocalSwarm` is used when refreshing endpoints after the first link |
| `nat`, `kcp`, `services` (opt) | NAT link strategy discovers NAT service, punches a UDP hole, then dials through KCP |
| `core/assets` | YAML config, endpoint database, loader setup, and migration |

## Flows

- Outbound link: strategy dials an `exonet.Conn` -> `noise.HandshakeOutbound` -> outbound negotiation chooses `mux2` and exchanges status and link nonce -> `newLink(outbound)` -> `LinkPool.AddLink` -> set mux router -> reflect observed endpoint.
- Inbound link: accepted connection -> `noise.HandshakeInbound` -> inbound negotiation offers `mux2`, status, and link nonce -> `newLink(inbound)` -> `LinkPool.AddLink` -> notify link watchers.
- Network route: reject non-network zone and self-target queries -> prefer existing non-high-pressure link -> otherwise `RetrieveLink` with basic and tor strategies for 120 seconds -> otherwise try `ExtraRelayVia` identities and push caller proof when caller differs from context identity.
- Session open: mux creates outbound session -> sends `Query` or `RelayQuery` -> waits for `Response` -> accepted response opens the session, grows writer credit, and copies response bytes back to the caller writer.
- Session accept: inbound `Query` or `RelayQuery` -> relayed queries require `Auth.Authorize(RelayForAction)` -> create inbound session -> wait for router assignment -> launch query in origin network zone -> send accepted or rejected `Response`.
- Frame loop: link read loop decodes frames into mux -> query frames run in goroutines -> `Data` frames push into input buffer -> `Read` frames grant writer credit -> `Reset` closes peer state.
- Flow control: `session.Write` chunks data at `MaxDataFrameSize` -> writer blocks when credit is exhausted -> peer `session.Read` sends `Read` credit after consuming bytes.
- Ping and pressure: link ping loop wakes after jitter, active checks, or strategy wake -> ping timeout closes the link -> pong RTT and bytes feed optional pressure detector.
- Session migration: initiator opens `nodes.migrate_session` -> pauses writer and swaps buffers -> ready and switched handshake -> old link sends migrate frame -> both sides resume and complete on the new link within `migrateSessionTimeout`.
- Connectivity upgrade: `ReceiveObject` sees `LinkPressureEvent` -> per-peer switch avoids overlapping upgrades -> prefer sibling link without pressure -> otherwise retrieve a new NAT link -> migrate eligible sessions -> apply cooldown.
- Strategy activation: `RetrieveLink` returns an existing link unless forced -> per-target `NodeLinker` signals requested strategies -> watcher receives the first successful link -> excess unconsumed links close with `ErrExcessLink`.
- Endpoint reflection and cleanup: links push observed endpoint messages when appropriate -> peer extracts public TCP or UTP endpoint -> bounded observed-endpoint cache emits `NewObservedEndpointEvent`; cleanup task removes expired endpoint rows after grace.

## Source

- `mod/nodes/module.go`, `mod/nodes/errors.go`, `mod/nodes/tasks.go`, `mod/nodes/relay_for_action.go`, `mod/nodes/migrate_signal.go`, `mod/nodes/endpoint_with_ttl.go`, `mod/nodes/link_info.go`, `mod/nodes/session_info.go`, `mod/nodes/node_info.go` - public API, constants, task interfaces, auth action, and info objects.
- `mod/nodes/frames/frame.go`, `mod/nodes/frames/ping.go`, `mod/nodes/frames/query.go`, `mod/nodes/frames/relay_query.go`, `mod/nodes/frames/read.go`, `mod/nodes/frames/response.go`, `mod/nodes/frames/data.go`, `mod/nodes/frames/migrate.go`, `mod/nodes/frames/reset.go`, `mod/nodes/frames/errors.go` - mux frame registry and wire encoders.
- `mod/nodes/src/loader.go`, `mod/nodes/src/deps.go`, `mod/nodes/src/module.go`, `mod/nodes/src/config.go` - config, database setup, strategy registration, resolver registration, and dependency hooks.
- `mod/nodes/src/link.go`, `mod/nodes/src/link_negotiator.go`, `mod/nodes/src/link_pool.go`, `mod/nodes/src/node_linker.go`, `mod/nodes/src/noise/handshake.go`, `mod/nodes/src/noise/conn.go` - link lifecycle, Noise handshake, negotiation, pooling, and strategy activation.
- `mod/nodes/src/basic_link_strategy.go`, `mod/nodes/src/tor_link_strategy.go`, `mod/nodes/src/nat_link_strategy.go` - dial race, tor retry and pressure behavior, NAT punch and KCP handoff.
- `mod/nodes/src/mux.go`, `mod/nodes/src/mux_sessions.go`, `mod/nodes/src/session.go`, `mod/nodes/src/mux_session_reader.go`, `mod/nodes/src/mux_session_writer.go`, `mod/nodes/src/input_buffer.go`, `mod/nodes/src/output_buffer.go`, `mod/nodes/src/session_migrator.go`, `mod/nodes/src/err_buffer_empty.go` - mux sessions, buffers, credit, and migration plumbing.
- `mod/nodes/src/query_router.go`, `mod/nodes/src/migrate_session.go`, `mod/nodes/src/connectivity_upgrade.go`, `mod/nodes/src/pressure.go` - network routing, migration orchestration, pressure-triggered upgrades, and pressure scoring.
- `mod/nodes/src/object_receiver.go`, `mod/nodes/src/object_describer.go`, `mod/nodes/src/object_finder.go`, `mod/nodes/src/ip_candidate_finder.go`, `mod/nodes/src/endpoint_resolver.go` - objects integration and endpoint discovery helpers.
- `mod/nodes/src/db.go`, `mod/nodes/src/db_endpoint.go`, `mod/nodes/src/db_endpoint_resolver.go`, `mod/nodes/src/cleanup_endpoints_task.go`, `mod/nodes/src/authorizers.go`, `mod/nodes/src/create_link_task.go`, `mod/nodes/src/ensure_link_task.go` - endpoint persistence, tasks, and relay authorizer.
- `mod/nodes/src/op_*.go` - query handlers for links, sessions, endpoints, link creation, close, and migration.
- `mod/nodes/client/client.go`, `mod/nodes/client/resolve_endpoints.go`, `mod/nodes/client/migrate_session.go` - typed clients used by strategies and migration flows.

## Surface

| What | Why it matters |
|---|---|
| `nodes.links`, `nodes.sessions`, `nodes.new_link`, `nodes.close_link`, `nodes.add_endpoint`, `nodes.resolve_endpoints`, `nodes.migrate_session` | external inspection and link/session control surface |
| `Module.RouteQuery` | network-zone router that chooses direct links, creates links, or routes through relays |
| `Mux.RouteQuery` | link-local router that opens a framed session for one in-flight query |
| `StrategyFactory`, `EndpointResolver` | extension points for link creation and endpoint discovery |
| `nodes__endpoints` | persistent endpoint table keyed by identity, network, and address with optional expiry |
| link and session events | objects/events integration used by discovery, pressure upgrades, and status reporting |

## Invariants

- `Mux.SetRouter` is one-shot; inbound queries wait on `routerSet` until the link is fully added.
- Inbound `AddLink` notifies watchers without a strategy name; outbound strategy success notifies with the strategy name.
- Session state transitions are compare-and-swap constrained: routing to open or closed, open to migrating, and any state to closed.
- Input buffer pushes are all-or-nothing; overflow returns `ErrBufferOverflow`.
- `Data` payloads are limited by `MaxDataFrameSize`, so writers must chunk larger writes.
- Auto-migration requires session age of at least 30 seconds or at least 1 MiB transferred; manual migration still requires an open session.
- `RelayForAction` grants only when `Actor` equals `ForID` and has no constraints.
- Per-target `NodeLinker` is single-instance and strategy signalling is idempotent while a strategy is active.
