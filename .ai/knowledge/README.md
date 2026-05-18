# Knowledge Index

Repo implementation knowledge. If this conflicts with code, trust code and update this.

## Concepts

Concept pages explain cross-module ideas. Read `concepts/README.md` before creating or reshaping concept files.

| Keywords | Read |
|---|---|
| Identity, secp256k1, Anyone, User, Swarm, node identity | `concepts/identity.md` |
| App, AppContract, GuestID, AccessToken, apphost IPC, Guest, handshake | `concepts/app.md` |
| user setup, node bootstrap, swarm join, invite flow, request invite, SignedNodeContract, active contract, first node | `concepts/user-setup.md` |
| Zone, ZoneNetwork, ZoneDevice, ZoneVirtual, network access, zone enforcement | `concepts/zone.md` |
| Query, Router, RouteQuery, RouteNotFound, Reject, Accept, Session, Preprocessor, Gateway, routing pipeline | `concepts/query.md` |
| Auth, Authorize, Action, ActionSudo, ActionRelayFor, auth handler, authorization | `concepts/auth.md` |
| lib/astrald, lib/apphost, lib/routing, lib/apps, lib/ipc, lib/query, astrald.Default, OpRouter, IncomingQuery, Serve, AppRegistrar, client library | `concepts/lib.md` |
| Object, ObjectID, Repository, Receiver, Describer, Searcher, Finder, Holder, objects.Load, objects.Save, repo group | `concepts/objects.md` |
| Op, operation, op_*.go, OpName, args struct, ops.Set, method name, query method string | `concepts/operations.md` |
| Node, module lifecycle, Load Inject LoadDependencies Prepare Run, Scheduler, core.Inject, core.Node | `concepts/node.md` |
| Transport, exonet, Stream, Link, link strategy, TCP, KCP, Tor, layer stack | `concepts/transport.md` |
| Link, LinkPool, LinkStrategy, LinkPressure, LinkCreatedEvent, LinkClosedEvent | `concepts/links.md` |
| Multiplexer, mux, Session, session migration, flow control, wsize, stateMigrating, Session vs Link | `concepts/mux.md` |
| brontide, Noise XK, handshake, forward secrecy, secp256k1 wire auth, RemotePub, authenticated connection | `concepts/brontide.md` |
| crypto, signing, Engine, EngineProvider, hash signing, text signing, BIP137, hardware wallet, Coldcard | `concepts/crypto.md` |
| Serialization, wire format, Objectify, WriteTo, ReadFrom, canonical encoding, ObjectType, astral primitives | `concepts/wire.md` |
| Channel, Switch, Handle, Collect, EOS, astral.Err, channel.Expect, channel helpers | `concepts/channels.md` |
| tree, tree.Value, Follow, live binding, settings, Mount, MountRemote, tree path, runtime config | `concepts/tree.md` |
| Presence, nearby, StatusMessage, Composer, Composition, stealth, broadcast, endpoint resolution | `concepts/presence.md` |

## Rules and Patterns

| Keywords | Read |
|---|---|
| Coding rule, constraint, invariant, style, naming, concurrency, mutex, atomic | `../rules.md` |
| Pattern, recipe, skeleton, boilerplate, how to write, module template, op handler | `../patterns/README.md` |

## Modules

Read the module guide when entering that module's source.

| Module path / keywords | Read |
|---|---|
| `mod/nodes/`, Link, Stream, Session, peer, flow control, frame protocol, migration, link establishment | `modules/nodes.md` |
| `mod/apphost/`, token, handler registration, IPC bridge, guest connection, contract indexing | `modules/apphost.md` |
| `mod/objects/`, Load[T], Save, Commit, Discard, Blueprint, repo group, Push, object store | `modules/objects.md` |
| `mod/dir/`, alias, filter, resolve, DisplayName, ApplyFilters, IdentityFilter, identity resolver | `modules/dir.md` |
| `mod/auth/`, Authorize, Add, auth handler | `modules/auth.md` |
| `mod/gateway/`, relay socket, binder, connector, gateway relay | `modules/gateway.md` |
| `mod/nat/`, hole punch, ConePuncher, UDP traversal, nat.Hole | `modules/nat.md` |
| `mod/kcp/`, KCP, UDP transport, local-port mapping, ephemeral listener | `modules/kcp.md` |
| `mod/tcp/`, TCP listener, ListenPort, CreateEphemeralListener | `modules/tcp.md` |
| `mod/exonet/`, Dialer, Unpacker, Parser, SetDialer, network name, transport registry | `modules/exonet.md` |
| `mod/user/`, user identity, Swarm member, MaintainLinkTask, node contract | `modules/user.md` |
| `mod/nearby/`, local discovery, broadcast, Stealth, Visible, UDP discovery | `modules/nearby.md` |
| `mod/scheduler/`, schedule task, run task, PoolLocker, Releaser, FuncAdapter | `modules/scheduler.md` |
| `mod/events/`, event, subscribe, emit, EventReceiver, EventEmitter | `modules/events.md` |
| `mod/fs/`, filesystem, file serve, ReadDir, Stat, virtual filesystem | `modules/fs.md` |
| `mod/ether/`, local network broadcast, UDP multicast, object broadcast, ether push | `modules/ether.md` |
| `mod/services/`, service registry, named service, bind service, AddService | `modules/services.md` |
| `mod/shell/`, shell command, terminal, admin CLI, command handler | `modules/shell.md` |
| `mod/tree/`, config tree, persistent setting, tree.Value, Follow, tree path | `modules/tree.md` |
| `mod/crypto/`, sign, verify, Engine, PrivateKey, SignableObject, secp256k1, BIP137 | `modules/crypto.md` |
| `mod/ip/`, LocalIPs, PublicIPCandidates, DefaultGateway, EventNetworkAddressChanged | `modules/ip.md` |
| `mod/tor/`, Tor, onion, hidden service, SOCKS5, ED25519-V3 | `modules/tor.md` |
| `mod/fwd/`, port forward, bridge, AstralServer, TCPServer, TorTarget | `modules/fwd.md` |
| `mod/log/`, logging, log level, OpListen, View, LogFile | `modules/log.md` |
