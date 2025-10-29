# Streams and Sessions in Astral's mod/nodes

## Overview

The `mod/nodes` module implements Astral's peer-to-peer networking layer through two core abstractions: **streams** and **sessions**. Streams represent authenticated, multiplexed transport connections between node identities. Sessions represent individual application-level data flows multiplexed over those streams. Together, they enable transparent migration of active connections across network paths while maintaining session continuity.

**Key Design Properties:**
- Streams multiplex multiple sessions over a single authenticated connection
- Sessions can migrate between streams without disrupting application data flow
- Migration is coordinated through a two-role FSM protocol with explicit signaling
- Frame-based protocol provides flow control and connection lifecycle management

## Architecture Layers

### Transport Layer: Streams

A `Stream` wraps an authenticated connection (`astral.Conn`) established through Noise protocol handshakes. Each stream:

- Has a unique local nonce identifier for migration targeting
- Belongs to exactly one pair of node identities (local and remote)
- Carries bidirectional frame traffic for all multiplexed sessions
- Maintains liveness through periodic ping/pong checks
- Terminates all attached sessions when closed

Streams are connection-oriented: they represent the underlying network path (TCP, KCP, UTP, etc.) wrapped in cryptographic authentication.

### Session Layer: Multiplexed Connections

A `session` represents an application-level connection multiplexed over a stream. Each session:

- Identified by a unique nonce shared between peers
- Implements `io.ReadWriteCloser` for application use
- Maintains independent read/write buffers with flow control
- Can transition between streams via migration without closing
- Has four states: `stateRouting`, `stateOpen`, `stateMigrating`, `stateClosed`

Sessions decouple application logic from transport paths, enabling seamless network transitions.

## Stream Lifecycle

### Creation and Authentication

**Outbound Connection:**
```
Connect(remoteID, conn) → noise handshake → feature negotiation → Stream
```

1. Caller initiates `Peers.Connect()` with target identity and raw connection
2. `noise.HandshakeOutbound()` authenticates remote identity using Noise protocol
3. Peer sends feature list and fresh stream nonce
4. Local node selects `mux2` feature, receives confirmation
5. `newStream()` creates Stream wrapper with nonce and outbound=true
6. Stream registered in `peers.streams` set

**Inbound Connection:**
```
Accept(conn) → noise handshake → feature negotiation → Stream
```

1. Listener calls `Peers.Accept()` with incoming connection
2. `noise.HandshakeInbound()` authenticates caller identity
3. Local node sends features and generates fresh stream nonce
4. Peer selects feature, receives confirmation
5. `newStream()` creates Stream wrapper with nonce and outbound=false
6. Stream registered and begins frame reading

### Frame Processing

Streams continuously read frames from the underlying connection:

```
readStreamFrames(s) → s.Read() → mod.in channel → frameReader()
```

The `frameReader()` goroutine dispatches frames by type:
- `Query`: Initiate new session
- `Response`: Accept/reject session establishment
- `Data`: Application payload delivery
- `Read`: Flow control advertisement
- `Reset`: Session termination
- `Ping`: Liveness check
- `Migrate`: Migration marker

### Termination

Stream closure cascades to all attached sessions:

1. Stream error detected or `CloseWithError()` called
2. Frame reader goroutine exits, closes channels
3. Stream removed from `peers.streams`
4. All sessions with `session.stream == s` are closed
5. `EventUnlinked` emitted if no other streams to same identity

## Session Lifecycle

### Establishment

**Outbound Query (Caller initiates):**

1. Application calls `node.RouteQuery()`
2. `Peers.RouteQuery()` creates new session in `stateRouting`
3. `frames.Query` sent on all streams to target identity
4. Session blocks waiting for `frames.Response` or context timeout
5. On accept: session transitions to `stateOpen`, returns to application
6. On reject: session deleted, error returned

**Inbound Query (Responder receives):**

1. `handleQuery()` receives `frames.Query` frame
2. New session created in `stateRouting` with query details
3. Query routed through local node's routing stack
4. On accept: `frames.Response` sent with buffer size, state → `stateOpen`
5. On reject: `frames.Response` sent with error code, session closed

### Data Transfer

Sessions implement standard `io.ReadWriteCloser` semantics:

**Write Path:**
- Application writes to session
- Session blocks if remote buffer full (flow control)
- Payload chunked to max 8192 bytes per `frames.Data`
- Frames sent on current `session.stream`
- Write advances only after successful frame transmission

**Read Path:**
- `frames.Data` arrives, queued to session's `rbuf`
- Application reads drain buffer
- Session sends `frames.Read` to advertise consumed buffer space
- Reads block when buffer empty and state is `stateOpen` or `stateMigrating`

### State Transitions

```
stateRouting → stateOpen:    Response accepted
stateOpen → stateMigrating:   Migration initiated
stateMigrating → stateOpen:   Migration completed
stateOpen → stateClosed:      Close() or Reset received
```

State changes broadcast to read/write condition variables to unblock waiting operations.

## Frame Protocol

### Frame Types

- **Query** (opcode 1): `{Nonce, Query, Buffer}` - Initiate session
- **Response** (opcode 2): `{Nonce, ErrCode, Buffer}` - Accept/reject
- **Data** (opcode 4): `{Nonce, Payload}` - Application data (≤8192 bytes)
- **Read** (opcode 3): `{Nonce, Len}` - Flow control credit
- **Reset** (opcode 5): `{Nonce}` - Terminate session
- **Ping** (opcode 0): `{Nonce, Pong}` - Liveness check
- **Migrate** (opcode 6): `{Nonce}` - Session migration marker

### Flow Control

Sessions implement window-based flow control:

- Each session advertises `rsize` (default 4MB) buffer capacity
- Sender tracks `wsize` (remote buffer space available)
- Writes block when `wsize == 0`
- Receiver sends `frames.Read{Len: n}` after consuming n bytes
- Sender increments `wsize` by n on receipt

This prevents sender overrun and provides backpressure.

## Migration Mechanism

### Purpose and Design

Migration enables transferring an active session from one stream to another without closing the application connection. Use cases include:
- Switching from high-latency to low-latency path
- Recovering from degraded connection quality
- Moving from cellular to WiFi network

The mechanism uses a two-role FSM protocol with out-of-band signaling over a separate query connection.

### Migration Roles

**Initiator:** The side that starts migration
- Sends BEGIN signal
- Waits for READY acknowledgment
- Calls `session.Migrate()` to enter `stateMigrating`
- Sends `frames.Migrate` marker on old stream
- Waits for COMPLETED signal
- Calls `session.CompleteMigration()` to switch stream and reopen

**Responder:** The side that receives migration request
- Receives BEGIN signal
- Calls `session.Migrate()` to enter `stateMigrating`
- Sends READY acknowledgment
- Waits for marker application (session reopened by peer's marker)
- Sends COMPLETED signal

### Signaling Protocol

Migration signaling occurs over a dedicated `nodes.migrate_session` query channel, separate from the session being migrated:

**FSM States:**
1. **StateMigrating**: Initial state, exchanging BEGIN/READY signals
2. **StateWaitingMarker**: Migration initiated, waiting for marker or completion
3. **StateCompleted**: Success
4. **StateFailed**: Error, abort sent

**Signal Types:**
- `MigrateSignalTypeBegin`: Initiator → Responder, start migration
- `MigrateSignalTypeReady`: Responder → Initiator, ready to proceed
- `MigrateSignalTypeCompleted`: Responder → Initiator, marker applied
- `MigrateSignalTypeAbort`: Either → Other, cancel migration

**Initiator Flow:**
```
Send BEGIN → Wait READY → Migrate() → Send Migrate frame → Wait COMPLETED → CompleteMigration()
```

**Responder Flow:**
```
Wait BEGIN → Migrate() → Send READY → Wait session reopen → Send COMPLETED
```

### Validation and Safety

- Target stream must connect to same remote identity as current stream
- Stream identified by local nonce via `Module.findStreamByID()`
- Migration FSM sends abort signal on any validation failure
- Both roles track session state atomically to prevent races
- Writes pause during migration (`stateMigrating` blocks in Write())

## Integration Points

### Peers Module

The `Peers` component orchestrates streams and sessions:

- **`streams sig.Set[*Stream]`**: Active streams registry
- **`sessions sig.Map[astral.Nonce, *session]`**: Active sessions by nonce
- **`RouteQuery()`**: Creates outbound sessions
- **`handleQuery()`**: Accepts inbound sessions
- **`addStream()`**: Registers new stream, starts frame reader
- **`frameReader()`**: Central frame dispatcher
- **`isLinked()`**: Checks if any stream exists to target identity

Events emitted:
- `EventLinked{NodeID}`: First stream to identity established
- `EventUnlinked{NodeID}`: Last stream to identity closed

### Module Interface

Public methods for migration management:

- **`OpMigrateSession()`**: CLI/RPC handler for migration requests
  - Accepts `{SessionID, StreamID, Start}` arguments
  - Instantiates `sessionMigrator` with role determination
  - Routes signaling channel to peer's `OpMigrateSession`
  - Runs FSM to completion

- **`findStreamByID(nonce)`**: Looks up stream by local identifier
  - Iterates `peers.streams` collection
  - Returns matching stream or nil
  - Used by migrator to resolve target stream

- **`CloseStream(nonce)`**: Terminates stream by identifier
  - Closes stream with error
  - Cascades to all attached sessions

### Objects Module

The migration signaling protocol uses the Objects module for out-of-band communication:

- Signaling messages are `astral.Object` types
- `SessionMigrateSignal` carries signal type and session nonce
- Transmitted over dedicated query channel separate from migrating session
- Channel created via `query.RouteChan()` using `nodes.migrate_session` method

### CLI and Diagnostics

Shell operations for manual migration:

- `OpMigrateSession`: Trigger migration between specific session/stream nonces
- Requires operator to identify session ID and target stream ID
- Primarily for testing and manual intervention scenarios

## Design Summary

### Key Architectural Decisions

**Separation of Transport and Session:**
- Streams handle authentication and framing
- Sessions handle application data and lifecycle
- Clean separation enables migration without application awareness

**Out-of-Band Signaling:**
- Migration coordination uses separate query channel
- Avoids complex in-band state machine on data stream
- Allows parallel signaling while session remains usable

**Explicit Marker Frame:**
- In-band `frames.Migrate` provides synchronization point
- Responder knows exactly when to apply migration
- Prevents ambiguity in data stream ordering

**State Machine Enforcement:**
- Session state atomically tracked via `atomic.Int32`
- Writes blocked during `stateMigrating` to prevent data loss
- State transitions broadcast to unblock waiting goroutines

**Flow Control:**
- Window-based protocol prevents buffer overrun
- Sender tracks remote space, blocks when full
- Receiver advertises consumption via Read frames

**Identity Preservation:**
- Sessions locked to remote identity, not stream
- Migration validated against identity mismatch
- Security properties maintained across network transitions

### Current limitations and future improvements

- Operator must manually identify session and target stream nonces
- No automatic migration triggers (quality degradation, path failure)
- Single stream failure terminates all sessions (no automatic recovery)